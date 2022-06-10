package dispatch

import (
	"context"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/dispatch/protocols"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/model/generic"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

type dispatchManager struct {
	config         *viper.Viper
	cloudConnector connectors.CloudConnectorClient
	db             *gorm.DB
	rateLimiter    *rate.Limiter
}

func (this *dispatchManager) newCorrelationId() uuid.UUID {
	if this.config.GetBool("demo.mode") {
		return uuid.UUID{}
	}

	return uuid.New()
}

func (this *dispatchManager) applyDefaults(run *generic.RunInput) {
	if run.WebConsoleUrl == nil {
		run.WebConsoleUrl = utils.StringRef(this.config.GetString("web.console.url.default"))
	}

	if run.Timeout == nil {
		run.Timeout = utils.IntRef(this.config.GetInt("default.run.timeout"))
	}
}

func getProtocol(runInput generic.RunInput) protocols.Protocol {
	if runInput.SatId != nil {
		return protocols.SatelliteProtocol
	} else {
		return protocols.RunnerProtocol
	}
}

func (this *dispatchManager) ProcessRun(ctx context.Context, account string, service string, run generic.RunInput) (runID, correlationID uuid.UUID, err error) {
	correlationID = this.newCorrelationId()
	ctx = utils.WithCorrelationId(ctx, correlationID.String())

	this.applyDefaults(&run)

	protocol := getProtocol(run)

	signalMetadata := protocol.BuildMedatada(run, correlationID, this.config)

	// take from the rate limit bucket
	rateErr := this.rateLimiter.Wait(ctx)

	if rateErr != nil {
		return uuid.UUID{}, correlationID, rateErr
	}

	messageId, notFound, err := this.cloudConnector.SendCloudConnectorRequest(
		ctx,
		account,
		run.Recipient,
		&run.Url,
		string(protocol.GetDirective()),
		signalMetadata,
	)

	if err != nil {
		instrumentation.CloudConnectorRequestError(ctx, err, run.Recipient, protocol.GetLabel())
		return uuid.UUID{}, correlationID, err
	} else if notFound {
		instrumentation.CloudConnectorNoConnection(ctx, run.Recipient, protocol.GetLabel())
		return uuid.UUID{}, correlationID, &RecipientNotFoundError{recipient: run.Recipient, err: err}
	}

	instrumentation.CloudConnectorOK(ctx, run.Recipient, messageId)

	entity := newRun(&run, correlationID, protocol.GetResponseFull(this.config), service, this.config)

	err = this.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if dbResult := tx.Create(&entity); dbResult.Error != nil {
			instrumentation.PlaybookRunCreateError(ctx, dbResult.Error, &entity, protocol.GetLabel())
			return dbResult.Error
		}

		if len(run.Hosts) > 0 {
			newHosts := newHostRun(run.Hosts, entity.ID)

			if dbResult := tx.Create(newHosts); dbResult.Error != nil {
				instrumentation.PlaybookRunHostCreateError(ctx, dbResult.Error, newHosts, protocol.GetLabel())
				return dbResult.Error
			}
		}

		return nil
	})

	if err != nil {
		return entity.ID, correlationID, err
	}

	instrumentation.RunCreated(ctx, run.Recipient, entity.ID, run.Url, entity.Service, protocol.GetLabel())
	return entity.ID, correlationID, nil
}

func (this *dispatchManager) ProcessCancel(ctx context.Context, account string, cancel generic.CancelInput) (runID, correlationID uuid.UUID, err error) {
	var run db.Run

	if err := this.db.First(&run, cancel.RunId).Error; err != nil {
		instrumentation.PlaybookRunCancelError(ctx, err)
		return uuid.UUID{}, run.CorrelationID, &RunNotFoundError{err: err, runID: cancel.RunId}
	}

	if run.SatId == nil || run.SatOrgId == nil {
		instrumentation.PlaybookRunCancelRunTypeError(ctx, run.ID)
		return uuid.UUID{}, run.CorrelationID, &RunCancelTypeError{err, run.ID}
	}

	if run.Status != db.RunStatusRunning {
		return uuid.UUID{}, run.CorrelationID, &RunCancelNotCancelableError{run.ID}
	}

	protocol := *protocols.SatelliteProtocol
	signalMetadata := protocol.BuildCancelMetadata(cancel, run.CorrelationID, this.config)

	// take from the rate limit bucket
	rateErr := this.rateLimiter.Wait(ctx)

	if rateErr != nil {
		return uuid.UUID{}, correlationID, rateErr
	}

	messageId, notFound, err := this.cloudConnector.SendCloudConnectorRequest(
		ctx,
		account,
		run.Recipient,
		nil,
		string(protocol.GetDirective()),
		signalMetadata,
	)

	if err != nil {
		instrumentation.CloudConnectorRequestError(ctx, err, run.Recipient, protocol.GetLabel())
		return uuid.UUID{}, run.CorrelationID, err
	} else if notFound {
		instrumentation.CloudConnectorNoConnection(ctx, run.Recipient, protocol.GetLabel())
		return uuid.UUID{}, run.CorrelationID, &RecipientNotFoundError{recipient: run.Recipient, err: err}
	}

	instrumentation.CloudConnectorOK(ctx, run.Recipient, messageId)
	instrumentation.RunCanceled(ctx, run.ID)

	return cancel.RunId, run.CorrelationID, nil
}
