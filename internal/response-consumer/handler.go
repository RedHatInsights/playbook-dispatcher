package responseConsumer

import (
	"context"
	"errors"
	"playbook-dispatcher/internal/common/ansible"
	"playbook-dispatcher/internal/common/constants"
	database "playbook-dispatcher/internal/common/db"
	kafkaUtils "playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/response-consumer/instrumentation"

	k "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	EventPlaybookOnStats  = "playbook_on_stats"
	EventRunnerOnFailed   = "runner_on_failed"
	EventExecutorOnFailed = "executor_on_failed"
)

type handler struct {
	db *gorm.DB
}

func (this *handler) onMessage(ctx context.Context, msg *k.Message) {
	requestId, correlationId, err := getHeaders(msg)
	if err != nil {
		instrumentation.CannotReadHeaders(ctx, err)
		return
	}

	ctx = utils.WithRequestId(ctx, requestId)
	ctx = utils.WithCorrelationId(ctx, correlationId.String())

	value := &message.PlaybookRunResponseMessageYaml{}

	if err := value.UnmarshalJSON(msg.Value); err != nil {
		instrumentation.UnmarshallIncomingMessageError(ctx, err)
		return
	}

	ctx = utils.WithAccount(ctx, value.Account)

	identity, err := utils.ParseIdentityHeader(value.B64Identity)
	if err == nil {
		ctx = utils.WithOrgId(ctx, identity.Identity.Internal.OrgID)
	}

	utils.GetLogFromContext(ctx).Debugw("Processing message",
		"upload_timestamp", value.UploadTimestamp,
		"topic", *msg.TopicPartition.Topic,
		"partition", msg.TopicPartition.Partition,
		"offset", msg.TopicPartition.Offset.String(),
	)

	status := inferStatus(&value.Events, nil)

	eventsSerialized := utils.MustMarshal(value.Events)

	var runsUpdated int64

	database.SetLog(this.db, utils.GetLogFromContext(ctx))
	defer database.ClearLog(this.db)

	run := db.Run{}

	err = this.db.Transaction(func(tx *gorm.DB) error {
		baseQuery := tx.Model(db.Run{}).
			Where("account = ?", value.Account).
			Where("correlation_id = ?", correlationId)

		if selectResult := baseQuery.Select("id").First(&run); selectResult.Error != nil {
			if errors.Is(selectResult.Error, gorm.ErrRecordNotFound) {
				return nil
			}

			utils.GetLogFromContext(ctx).Errorw("Error fetching run from db", "error", selectResult.Error)
			return selectResult.Error
		}

		toUpdate := db.Run{
			Status: status,
			Events: eventsSerialized,
		}

		if updateResult := baseQuery.Select("status", "events").Updates(toUpdate); updateResult.Error != nil {
			utils.GetLogFromContext(ctx).Errorw("Error updating run in db", "error", updateResult.Error)
			return updateResult.Error
		} else {
			runsUpdated = updateResult.RowsAffected
		}

		hosts := ansible.GetAnsibleHosts(value.Events)

		if len(hosts) == 0 {
			return nil
		}

		toCreate := mapHostsToRunHosts(hosts, func(host string) db.RunHost {
			return db.RunHost{
				ID:     uuid.New(),
				RunID:  run.ID,
				Host:   host,
				Status: inferStatus(&value.Events, &host),
				Log:    ansible.GetStdout(value.Events, nil),
			}
		})

		createResult := tx.Model(db.RunHost{}).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "run_id"}, {Name: "host"}},
				DoUpdates: clause.AssignmentColumns([]string{"status", "log"}),
			}).
			Create(&toCreate)

		if createResult.Error != nil {
			utils.GetLogFromContext(ctx).Errorw("Error upserting run hosts in db", "error", createResult.Error)
			return createResult.Error
		}

		return nil
	})

	if err != nil {
		instrumentation.PlaybookRunUpdateError(ctx, err, status, run.ID)
	} else if runsUpdated > 0 {
		instrumentation.PlaybookRunUpdated(ctx, status, run.ID)
	} else {
		instrumentation.PlaybookRunUpdateMiss(ctx, status)
	}
}

func inferStatus(events *[]message.PlaybookRunResponseMessageYamlEventsElem, host *string) string {
	finished := false
	failed := false

	for _, event := range *events {
		if event.Event == EventPlaybookOnStats {
			finished = true
		}

		// if host parameter is defined only consider events for the given host
		if host != nil && event.EventData != nil && event.EventData.Host != nil && *event.EventData.Host != *host {
			continue
		}

		if event.Event == EventRunnerOnFailed {
			failed = true
		}

		if event.Event == EventExecutorOnFailed {
			failed = true
			finished = true
		}
	}

	switch {
	case finished && failed:
		return db.RunStatusFailure
	case finished && !failed:
		return db.RunStatusSuccess
	default:
		return db.RunStatusRunning
	}
}

func getHeaders(msg *k.Message) (requestId string, correlationId uuid.UUID, err error) {
	if requestId, err = kafkaUtils.GetHeader(msg, constants.HeaderRequestId); err != nil {
		return
	}

	var correlationIdRaw string
	if correlationIdRaw, err = kafkaUtils.GetHeader(msg, constants.HeaderCorrelationId); err != nil {
		return
	}

	if correlationId, err = uuid.Parse(correlationIdRaw); err != nil {
		return
	}

	return
}

func mapHostsToRunHosts(hosts []string, fn func(host string) db.RunHost) []db.RunHost {
	result := make([]db.RunHost, len(hosts))
	for i, host := range hosts {
		result[i] = fn(host)
	}

	return result
}
