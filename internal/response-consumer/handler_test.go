package responseConsumer

import (
	"encoding/json"
	"playbook-dispatcher/internal/common/constants"
	kafkaUtils "playbook-dispatcher/internal/common/kafka"
	dbModel "playbook-dispatcher/internal/common/model/db"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"
	"sort"

	k "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var localhost = "localhost"

func newResponseMessage(events *[]messageModel.PlaybookRunResponseMessageYamlEventsElem, correlationId uuid.UUID) *k.Message {
	data := messageModel.PlaybookRunResponseMessageYaml{
		Account:   accountNumber(),
		RequestId: uuid.New().String(),
		Events:    *events,
	}

	value, err := json.Marshal(data)
	Expect(err).ToNot(HaveOccurred())

	topic := "platform.playbook-dispatcher.runs"

	return &k.Message{
		Value:   value,
		Headers: kafkaUtils.Headers(constants.HeaderCorrelationId, correlationId.String(), constants.HeaderRequestId, "test"),
		TopicPartition: k.TopicPartition{
			Topic:     &topic,
			Partition: 0,
			Offset:    k.Offset(0),
		},
	}
}

func createEvents(events ...string) *[]messageModel.PlaybookRunResponseMessageYamlEventsElem {
	result := make([]messageModel.PlaybookRunResponseMessageYamlEventsElem, len(events))

	for i, event := range events {
		result[i] = messageModel.PlaybookRunResponseMessageYamlEventsElem{
			Counter: i,
			Event:   event,
			EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
				Host: &localhost,
			},
		}
	}

	return &result
}

var _ = Describe("handler", func() {
	var instance handler

	db := test.WithDatabase()

	BeforeEach(func() {
		instance = handler{
			db: db(),
		}
	})

	fetchRun := func(id uuid.UUID) *dbModel.Run {
		var run dbModel.Run
		result := db().Where("id = ?", id.String()).First(&run)
		Expect(result.Error).ToNot(HaveOccurred())
		return &run
	}

	fetchHosts := func(runId uuid.UUID) []dbModel.RunHost {
		var hosts []dbModel.RunHost
		result := db().Table("run_hosts").Where("run_id = ?", runId.String()).Find(&hosts)
		Expect(result.Error).ToNot(HaveOccurred())
		return hosts
	}

	checkHost := func(runId uuid.UUID, status string, log string) {
		hosts := fetchHosts(runId)
		Expect(hosts).To(HaveLen(1))
		Expect(hosts[0].RunID).To(Equal(runId))
		Expect(hosts[0].InventoryID).To(BeNil())
		Expect(hosts[0].Host).To(Equal("localhost"))
		Expect(hosts[0].Status).To(Equal(status))
		Expect(hosts[0].Log).To(Equal(log))
	}

	Describe("state update", func() {
		It("noop on empty list of events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			msg := newResponseMessage(&[]messageModel.PlaybookRunResponseMessageYamlEventsElem{}, uuid.New())
			instance.onMessage(test.TestContext(), msg)

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("running"))

			hosts := fetchHosts(data.ID)
			Expect(hosts).To(BeEmpty())
		})

		It("updates the run status based on successful runner events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createEvents(
				messageModel.EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_ok",
				"playbook_on_stats",
			)

			instance.onMessage(test.TestContext(), newResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("success"))
			checkHost(data.ID, "success", "")
		})

		It("updates the run status based on failure runner events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createEvents(
				messageModel.EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_failed",
				"playbook_on_stats",
			)

			instance.onMessage(test.TestContext(), newResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", "")
		})

		It("updates multiple hosts involved in a run", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createEvents(
				messageModel.EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start", // host 1
				"runner_on_start", // host 2
				"runner_on_ok",    // host 1
				"runner_on_ok",    // host 2
				"playbook_on_task_start",
				"runner_on_start",  // host 1
				"runner_on_start",  // host 2
				"runner_on_ok",     // host 1
				"runner_on_failed", // host 2
				"playbook_on_stats",
			)

			localhost2 := "localhost2"
			(*events)[5].EventData.Host = &localhost2
			(*events)[6].Stdout = utils.StringRef("a")
			(*events)[7].EventData.Host = &localhost2
			(*events)[7].Stdout = utils.StringRef("1")
			(*events)[10].EventData.Host = &localhost2
			(*events)[11].Stdout = utils.StringRef("b")
			(*events)[12].EventData.Host = &localhost2
			(*events)[12].Stdout = utils.StringRef("2")

			instance.onMessage(test.TestContext(), newResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("failure"))
			hosts := fetchHosts(data.ID)
			Expect(hosts).To(HaveLen(2))

			sort.Slice(hosts, func(i, j int) bool {
				return hosts[i].Host == "localhost"
			})

			Expect(hosts[0].RunID).To(Equal(data.ID))
			Expect(hosts[1].RunID).To(Equal(data.ID))
			Expect(hosts[0].InventoryID).To(BeNil())
			Expect(hosts[1].InventoryID).To(BeNil())
			Expect(hosts[0].Host).To(Equal("localhost"))
			Expect(hosts[1].Host).To(Equal("localhost2"))
			Expect(hosts[0].Status).To(Equal("success"))
			Expect(hosts[1].Status).To(Equal("failure"))
			Expect(hosts[0].Log).To(Equal("a1b2"))
			Expect(hosts[1].Log).To(Equal("a1b2"))
		})
	})

	Describe("correlation", func() {
		It("updates the correct run", func() {
			data := []dbModel.Run{
				test.NewRun(accountNumber()),
				test.NewRun(accountNumber()),
			}

			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createEvents(
				messageModel.EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_ok",
				"playbook_on_stats",
			)

			msg := newResponseMessage(events, data[1].CorrelationID)
			instance.onMessage(test.TestContext(), msg)

			run0 := fetchRun(data[0].ID)
			Expect(run0.Status).To(Equal("running"))
			hosts := fetchHosts(data[0].ID)
			Expect(hosts).To(BeEmpty())

			run1 := fetchRun(data[1].ID)
			Expect(run1.Status).To(Equal("success"))
			checkHost(data[1].ID, "success", "")
		})
	})
})
