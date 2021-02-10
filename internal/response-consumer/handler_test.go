package responseConsumer

import (
	"encoding/json"
	dbModel "playbook-dispatcher/internal/common/model/db"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils/test"

	"go.uber.org/zap"

	k "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newResponseMessage(events *[]messageModel.PlaybookRunResponseMessageYamlEventsElem) *k.Message {
	data := messageModel.PlaybookRunResponseMessageYaml{
		Account:   accountNumber(),
		RequestId: uuid.New().String(),
		Events:    *events,
	}

	value, err := json.Marshal(data)
	Expect(err).ToNot(HaveOccurred())

	return &k.Message{Value: value}
}

func createEvents(events ...string) *[]messageModel.PlaybookRunResponseMessageYamlEventsElem {
	result := make([]messageModel.PlaybookRunResponseMessageYamlEventsElem, len(events))

	for i, event := range events {
		result[i] = messageModel.PlaybookRunResponseMessageYamlEventsElem{
			Counter: i,
			Event:   event,
		}
	}

	return &result
}

var _ = Describe("handler", func() {
	var instance handler

	db := test.WithDatabase()

	BeforeEach(func() {
		instance = handler{
			db:  db(),
			log: zap.NewNop().Sugar(),
		}
	})

	fetchRun := func(id uuid.UUID) *dbModel.Run {
		var run dbModel.Run
		result := db().Where("id = ?", id.String()).First(&run)
		Expect(result.Error).ToNot(HaveOccurred())
		return &run
	}

	Describe("state update", func() {
		It("noop on empty list of events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			msg := newResponseMessage(&[]messageModel.PlaybookRunResponseMessageYamlEventsElem{})
			instance.onMessage(msg)

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("running"))
		})

		It("updates the run status based on successful runner events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createEvents(
				EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_ok",
				"playbook_on_stats",
			)

			correlationId := data.CorrelationID.String()
			(*events)[0].EventData = &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
				CrcCorrelationId: &correlationId,
			}

			instance.onMessage(newResponseMessage(events))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("success"))
		})

		It("updates the run status based on failure runner events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createEvents(
				EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_failed",
				"playbook_on_stats",
			)

			correlationId := data.CorrelationID.String()
			(*events)[0].EventData = &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
				CrcCorrelationId: &correlationId,
			}

			instance.onMessage(newResponseMessage(events))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("failure"))
		})
	})

	Describe("correlation", func() {
		It("updates the correct run", func() {
			data := []*dbModel.Run{
				test.NewRun(accountNumber()),
				test.NewRun(accountNumber()),
			}

			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createEvents(
				EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_ok",
				"playbook_on_stats",
			)

			run1CorrelationId := data[1].CorrelationID.String()

			(*events)[0].EventData = &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
				CrcCorrelationId: &run1CorrelationId,
			}

			msg := newResponseMessage(events)
			instance.onMessage(msg)

			run0 := fetchRun(data[0].ID)
			Expect(run0.Status).To(Equal("running"))

			run1 := fetchRun(data[1].ID)
			Expect(run1.Status).To(Equal("success"))
		})
	})
})
