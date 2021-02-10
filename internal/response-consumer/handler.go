package responseConsumer

import (
	"context"
	"fmt"
	"playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/response-consumer/instrumentation"

	"go.uber.org/zap"

	k "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	EventPlaybookOnStats = "playbook_on_stats"
	EventRunnerOnFailed  = "runner_on_failed"
	EventExecutorOnStart = "executor_on_start"
)

type handler struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

func (this *handler) onMessage(msg *k.Message) {
	value := &message.PlaybookRunResponseMessageYaml{}

	if err := value.UnmarshalJSON(msg.Value); err != nil {
		instrumentation.UnmarshallIncomingMessageError(this.log, err)
		return
	}

	// TODO: get ctx as onMessage param
	ctx, log := utils.WithRequestId(utils.SetLog(context.Background(), this.log), value.RequestId)
	log.Debugw("Processing message", "account", value.Account, "upload_timestamp", value.UploadTimestamp)

	status := inferStatus(&value.Events)

	queryBuilder := this.db.Model(db.Run{}).
		Select("status").
		Where("account = ?", value.Account)

	// TODO: validator should be infering this and setting it as a message key
	correlationId, err := inferCorrelationId(&value.Events)
	if err != nil {
		instrumentation.InvalidCorrelationId(ctx, err)
		return
	} else {
		queryBuilder.Where("correlation_id = ?", correlationId)
	}

	result := queryBuilder.Updates(db.Run{Status: status})

	if result.Error != nil {
		instrumentation.PlaybookRunUpdateError(ctx, result.Error, value.Account, status, correlationId)
	} else if result.RowsAffected > 0 {
		instrumentation.PlaybookRunUpdated(ctx, value.Account, status, correlationId)
	} else {
		instrumentation.PlaybookRunUpdateMiss(ctx, value.Account, status, correlationId)
	}
}

func inferStatus(events *[]message.PlaybookRunResponseMessageYamlEventsElem) string {
	finished := false
	failed := false

	for _, event := range *events {
		if event.Event == EventPlaybookOnStats {
			finished = true
		}

		if event.Event == EventRunnerOnFailed {
			failed = true
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

func inferCorrelationId(events *[]message.PlaybookRunResponseMessageYamlEventsElem) (result uuid.UUID, err error) {
	for _, event := range *events {
		if event.Event == EventExecutorOnStart && event.EventData != nil && event.EventData.CrcCorrelationId != nil {
			result, err = uuid.Parse(*event.EventData.CrcCorrelationId)
			return
		}
	}

	err = fmt.Errorf("Correlation id not found")
	return
}
