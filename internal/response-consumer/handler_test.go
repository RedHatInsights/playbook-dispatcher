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

func newResponseMessage(value interface{}, correlationId uuid.UUID, requestType string) *k.Message {
	marshalled, err := json.Marshal(value)
	Expect(err).ToNot(HaveOccurred())

	topic := "platform.playbook-dispatcher.runs"

	return &k.Message{
		Value:   marshalled,
		Headers: kafkaUtils.Headers(constants.HeaderCorrelationId, correlationId.String(), constants.HeaderRequestId, "test", constants.HeaderRequestType, requestType),
		TopicPartition: k.TopicPartition{
			Topic:     &topic,
			Partition: 0,
			Offset:    k.Offset(0),
		},
	}
}

func newSatResponseMessage(satEvents *[]messageModel.PlaybookSatRunResponseMessageYamlEventsElem, correlationId uuid.UUID) *k.Message {
	data := messageModel.PlaybookSatRunResponseMessageYaml{
		Account:   accountNumber(),
		RequestId: uuid.New().String(),
		Events:    *satEvents,
	}

	return newResponseMessage(data, correlationId, satMessageHeaderValue)
}

func newRunnerResponseMessage(runnerEvents *[]messageModel.PlaybookRunResponseMessageYamlEventsElem, correlationId uuid.UUID) *k.Message {
	data := messageModel.PlaybookRunResponseMessageYaml{
		Account:   accountNumber(),
		RequestId: uuid.New().String(),
		Events:    *runnerEvents,
	}

	return newResponseMessage(data, correlationId, runnerMessageHeaderValue)
}

func createRunnerEvents(events ...string) *[]messageModel.PlaybookRunResponseMessageYamlEventsElem {
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

type eventData struct {
	Type   string
	Status string
}

func createSatEvents(correlationId uuid.UUID, events ...eventData) *[]messageModel.PlaybookSatRunResponseMessageYamlEventsElem {
	result := make([]messageModel.PlaybookSatRunResponseMessageYamlEventsElem, len(events))

	for i, event := range events {
		status := messageModel.PlaybookSatRunResponseMessageYamlEventsElemStatus(event.Status)
		result[i] = messageModel.PlaybookSatRunResponseMessageYamlEventsElem{
			Type:          messageModel.PlaybookSatRunResponseMessageYamlEventsElemType(event.Type),
			Version:       3,
			Sequence:      &i,
			CorrelationId: correlationId.String(),
			Host:          &localhost,
			Status:        &status,
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

	checkHost := func(runId uuid.UUID, status string, satSeq *int, log string) {
		hosts := fetchHosts(runId)
		Expect(hosts).To(HaveLen(1))
		Expect(hosts[0].RunID).To(Equal(runId))
		Expect(hosts[0].InventoryID).To(BeNil())
		Expect(hosts[0].Host).To(Equal("localhost"))
		Expect(hosts[0].Status).To(Equal(status))
		Expect(hosts[0].SatSequence).To(Equal(satSeq))
		Expect(hosts[0].Log).To(Equal(log))
	}

	Describe("state update", func() {
		It("noop on empty list of events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			msg := newRunnerResponseMessage(&[]messageModel.PlaybookRunResponseMessageYamlEventsElem{}, uuid.New())
			instance.onMessage(test.TestContext(), msg)

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("running"))

			hosts := fetchHosts(data.ID)
			Expect(hosts).To(BeEmpty())
		})

		It("updates the run status based on successful runner events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createRunnerEvents(
				messageModel.EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_ok",
				"playbook_on_stats",
			)

			instance.onMessage(test.TestContext(), newRunnerResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("success"))
			checkHost(data.ID, "success", nil, "")
		})

		It("updates the run status based on failure runner events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createRunnerEvents(
				messageModel.EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_failed",
				"playbook_on_stats",
			)

			instance.onMessage(test.TestContext(), newRunnerResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", nil, "")
		})

		It("updates multiple hosts involved in a run", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createRunnerEvents(
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

			instance.onMessage(test.TestContext(), newRunnerResponseMessage(events, data.CorrelationID))

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

		It("updates the run status based on successful satellite events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createSatEvents(
				data.CorrelationID,
				eventData{"playbook_run_update", "success"},
				eventData{"playbook_run_update", "success"},
				eventData{"playbook_run_finished", "success"},
			)

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 2
			Expect(run.Status).To(Equal("success"))
			checkHost(data.ID, "success", &seq, "")
		})

		It("updates the run status based on failed satellite events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createSatEvents(
				data.CorrelationID,
				eventData{"playbook_run_update", "success"},
				eventData{"playbook_run_update", "failure"},
				eventData{"playbook_run_finished", "failure"},
			)

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 2
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", &seq, "")
		})

		It("updates the run status based on canceled satellite events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createSatEvents(
				data.CorrelationID,
				eventData{"playbook_run_update", "success"},
				eventData{"playbook_run_finished", "canceled"},
			)

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 1
			Expect(run.Status).To(Equal("canceled"))
			checkHost(data.ID, "canceled", &seq, "")
		})

		It("updates multiple satellite hosts involved in a run", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createSatEvents(
				data.CorrelationID,
				eventData{"playbook_run_update", "success"},   // host 1
				eventData{"playbook_run_update", "success"},   // host 2
				eventData{"playbook_run_finished", "success"}, //host 1
				eventData{"playbook_run_finished", "failure"}, //host 2
			)

			localhost2 := "localhost2"
			(*events)[0].Console = utils.StringRef("c3")
			(*events)[1].Host = &localhost2
			(*events)[1].Console = utils.StringRef("e5")
			(*events)[2].Console = utils.StringRef("d4")
			(*events)[3].Host = &localhost2
			(*events)[3].Console = utils.StringRef("f6")

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			Expect(run.Status).To(Equal("failure"))
			hosts := fetchHosts(data.ID)
			Expect(hosts).To(HaveLen(2))

			sort.Slice(hosts, func(i, j int) bool {
				return hosts[i].Host == "localhost"
			})

			Expect(hosts[0].RunID).To(Equal(data.ID))
			Expect(hosts[1].RunID).To(Equal(data.ID))
			Expect(hosts[0].Host).To(Equal("localhost"))
			Expect(hosts[1].Host).To(Equal("localhost2"))
			Expect(hosts[0].Status).To(Equal("success"))
			Expect(hosts[1].Status).To(Equal("failure"))
			Expect(hosts[0].Log).To(Equal("c3d4"))
			Expect(hosts[1].Log).To(Equal("e5f6"))
		})

		It("copies over satellite_connection_error to console", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createSatEvents(
				data.CorrelationID,
				eventData{"playbook_run_update", "success"},
				eventData{"playbook_run_finished", "failure"},
			)

			code := 1
			errorDescription := "Satellite unreachable"

			(*events)[1].SatelliteConnectionCode = &code
			(*events)[1].SatelliteConnectionError = &errorDescription

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 1
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", &seq, errorDescription)
		})

		It("copies over satellite_infrastructure_error to console", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createSatEvents(
				data.CorrelationID,
				eventData{"playbook_run_update", "success"},
				eventData{"playbook_run_finished", "failure"},
			)

			code := 1
			errorDescription := "Capsule is down"

			(*events)[1].SatelliteInfrastructureCode = &code
			(*events)[1].SatelliteInfrastructureError = &errorDescription

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 1
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", &seq, errorDescription)
		})
	})

	Describe("correlation", func() {
		It("updates the correct run", func() {
			data := []dbModel.Run{
				test.NewRun(accountNumber()),
				test.NewRun(accountNumber()),
			}

			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			events := createRunnerEvents(
				messageModel.EventExecutorOnStart,
				"playbook_on_start",
				"playbook_on_play_start",
				"playbook_on_task_start",
				"runner_on_start",
				"runner_on_ok",
				"playbook_on_stats",
			)

			msg := newRunnerResponseMessage(events, data[1].CorrelationID)
			instance.onMessage(test.TestContext(), msg)

			run0 := fetchRun(data[0].ID)
			Expect(run0.Status).To(Equal("running"))
			hosts := fetchHosts(data[0].ID)
			Expect(hosts).To(BeEmpty())

			run1 := fetchRun(data[1].ID)
			Expect(run1.Status).To(Equal("success"))
			checkHost(data[1].ID, "success", nil, "")
		})
	})
})
