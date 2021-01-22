package responseConsumer

import (
	"playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"

	"go.uber.org/zap"

	k "github.com/confluentinc/confluent-kafka-go/kafka"
	"gorm.io/gorm"
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

	result := this.db.Model(db.Run{}).
		Select("status").
		Where("account = ?", value.Account). // TODO: correlation id
		Updates(db.Run{Status: status})

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
		if event.Event == "playbook_on_stats" {
			finished = true
		}

		if event.Event == "runner_on_failed" {
			failed = true
		}
	}

	switch {
	case finished && failed:
		return "failure" // TODO constants
	case finished && !failed:
		return "success"
	default:
		return "running"
	}
}
