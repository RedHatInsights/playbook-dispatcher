package responseConsumer

import (
	"playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"

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
		this.log.Error(err)
		return
	}

	log := utils.LogWithRequestId(this.log, value.RequestId)
	log.Infow("Processing message", "account", value.Account, "upload_timestamp", value.UploadTimestamp)

	status := inferStatus(&value.Events)

	queryBuilder := this.db.Model(db.Run{}).
		Select("status").
		Where("account = ?", value.Account)

	if correlationId, err := inferCorrelationId(&value.Events); err != nil {
		log.Error("Failed to parse correlation id", "error", err)
	} else if correlationId != nil {
		queryBuilder.Where("correlation_id = ?", *correlationId)
	}

	result := queryBuilder.Updates(db.Run{Status: status})

	if result.Error != nil {
		log.Error(result.Error)
	} else {
		log.Infow("Updated run", "account", value.Account, "status", status, "count", result.RowsAffected)
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

func inferCorrelationId(events *[]message.PlaybookRunResponseMessageYamlEventsElem) (result *uuid.UUID, err error) {
	for _, event := range *events {
		if event.Event == EventExecutorOnStart && event.EventData != nil && event.EventData.CrcCorrelationId != nil {
			correlationId, err := uuid.Parse(*event.EventData.CrcCorrelationId)
			return &correlationId, err
		}
	}

	return nil, nil
}
