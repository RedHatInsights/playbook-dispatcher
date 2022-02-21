package dispatch

import (
	"context"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/dispatch/protocols"
	"playbook-dispatcher/internal/api/instrumentation"
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
	this.rateLimiter.Wait(ctx)

	messageId, notFound, err := this.cloudConnector.SendCloudConnectorRequest(
		ctx,
		account,
		run.Recipient,
		run.Url,
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
