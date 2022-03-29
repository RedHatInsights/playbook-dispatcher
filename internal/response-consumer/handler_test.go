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

type satEventBuilder func(correlationId string) messageModel.PlaybookSatRunResponseMessageYamlEventsElem

func satPlaybookRunUpdateEvent(sequence int, host, console string) satEventBuilder {
	return func(correlationId string) messageModel.PlaybookSatRunResponseMessageYamlEventsElem {
		return messageModel.PlaybookSatRunResponseMessageYamlEventsElem{
			Type:          messageModel.PlaybookSatRunResponseMessageYamlEventsElemTypePlaybookRunUpdate,
			Version:       3,
			CorrelationId: correlationId,
			Sequence:      &sequence,
			Host:          &host,
			Console:       &console,
		}
	}
}

func satPlaybookRunFinishedEvent(host, status string) satEventBuilder {
	return func(correlationId string) messageModel.PlaybookSatRunResponseMessageYamlEventsElem {
		status := messageModel.PlaybookSatRunResponseMessageYamlEventsElemStatus(status)
		return messageModel.PlaybookSatRunResponseMessageYamlEventsElem{
			Type:           messageModel.PlaybookSatRunResponseMessageYamlEventsElemTypePlaybookRunFinished,
			Version:        3,
			CorrelationId:  correlationId,
			Host:           &host,
			Status:         &status,
			ExecutionCode:  utils.IntRef(0),
			ConnectionCode: utils.IntRef(0),
		}
	}
}

func satPlaybookRunCompletedEvent(status string) satEventBuilder {
	return func(correlationId string) messageModel.PlaybookSatRunResponseMessageYamlEventsElem {
		status := messageModel.PlaybookSatRunResponseMessageYamlEventsElemStatus(status)
		return messageModel.PlaybookSatRunResponseMessageYamlEventsElem{
			Type:                         messageModel.PlaybookSatRunResponseMessageYamlEventsElemTypePlaybookRunCompleted,
			Version:                      3,
			CorrelationId:                correlationId,
			Status:                       &status,
			SatelliteConnectionCode:      utils.IntRef(0),
			SatelliteInfrastructureCode:  utils.IntRef(0),
			SatelliteConnectionError:     nil,
			SatelliteInfrastructureError: nil,
		}
	}
}

func buildSatEvents(correlationId uuid.UUID, builders ...satEventBuilder) *[]messageModel.PlaybookSatRunResponseMessageYamlEventsElem {
	result := make([]messageModel.PlaybookSatRunResponseMessageYamlEventsElem, len(builders))

	for i, builder := range builders {
		result[i] = builder(correlationId.String())
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

	checkHost := func(runId uuid.UUID, status string, satSeq *int, log string, inventoryId *uuid.UUID) {
		hosts := fetchHosts(runId)
		Expect(hosts).To(HaveLen(1))
		Expect(hosts[0].RunID).To(Equal(runId))
		Expect(hosts[0].InventoryID).To(Equal(inventoryId))
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
			checkHost(data.ID, "success", nil, "", nil)
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
			checkHost(data.ID, "failure", nil, "", nil)
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
			checkHost(data[1].ID, "success", nil, "", nil)
		})
	})

	Describe("Satellite", func() {
		It("updates the run status based on successful satellite events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			inventoryId := uuid.New()
			var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
			inventoryIdString := inventoryId.String()

			Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

			events := buildSatEvents(
				data.CorrelationID,
				satPlaybookRunUpdateEvent(1, inventoryIdString, ""),
				satPlaybookRunUpdateEvent(2, inventoryIdString, ""),
				satPlaybookRunFinishedEvent(inventoryIdString, "success"),
			)

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 2
			Expect(run.Status).To(Equal("success"))
			checkHost(data.ID, "success", &seq, "", &inventoryId)
		})

		It("updates the run status based on failed satellite events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			inventoryId := uuid.New()
			var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
			inventoryIdString := inventoryId.String()

			Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

			events := buildSatEvents(
				data.CorrelationID,
				satPlaybookRunUpdateEvent(1, inventoryIdString, ""),
				satPlaybookRunUpdateEvent(2, inventoryIdString, ""),
				satPlaybookRunFinishedEvent(inventoryIdString, "failure"),
			)

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 2
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", &seq, "", &inventoryId)
		})

		It("updates the run status based on canceled satellite events", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			inventoryId := uuid.New()
			var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
			inventoryIdString := inventoryId.String()

			Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

			events := buildSatEvents(
				data.CorrelationID,
				satPlaybookRunUpdateEvent(1, inventoryIdString, ""),
				satPlaybookRunFinishedEvent(inventoryIdString, "canceled"),
			)

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 1
			Expect(run.Status).To(Equal("canceled"))
			checkHost(data.ID, "canceled", &seq, "", &inventoryId)
		})

		It("updates multiple satellite hosts involved in a run", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			inventoryId := uuid.New()
			var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
			inventoryId1String := inventoryId.String()

			Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

			inventoryId2 := uuid.New()
			var host2Data = test.NewRunHost(data.ID, "running", &inventoryId2)
			host2Data.Host = "localhost2"
			inventoryId2String := inventoryId2.String()

			Expect(db().Create(&host2Data).Error).ToNot(HaveOccurred())

			events := buildSatEvents(
				data.CorrelationID,
				satPlaybookRunUpdateEvent(0, inventoryId1String, "c3"),
				satPlaybookRunUpdateEvent(0, inventoryId2String, "e5"),
				satPlaybookRunUpdateEvent(1, inventoryId1String, "d4"),
				satPlaybookRunUpdateEvent(1, inventoryId2String, "f6"),
				satPlaybookRunFinishedEvent(inventoryId1String, "success"),
				satPlaybookRunFinishedEvent(inventoryId2String, "failure"),
			)

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
			Expect(hosts[0].InventoryID).To(Equal(&inventoryId))
			Expect(hosts[1].InventoryID).To(Equal(&inventoryId2))
			Expect(hosts[0].Status).To(Equal("success"))
			Expect(hosts[1].Status).To(Equal("failure"))
			Expect(hosts[0].Log).To(Equal("c3d4"))
			Expect(hosts[1].Log).To(Equal("e5f6"))
		})

		It("copies over satellite_connection_error to console", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			inventoryId := uuid.New()
			var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
			inventoryIdString := inventoryId.String()

			Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

			events := buildSatEvents(
				data.CorrelationID,
				satPlaybookRunUpdateEvent(1, inventoryIdString, ""),
				satPlaybookRunFinishedEvent(inventoryIdString, "failure"),
			)

			code := 1
			errorDescription := "Satellite unreachable"

			(*events)[1].SatelliteConnectionCode = &code
			(*events)[1].SatelliteConnectionError = &errorDescription

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 1
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", &seq, errorDescription, &inventoryId)
		})

		It("copies over satellite_infrastructure_error to console", func() {
			var data = test.NewRun(accountNumber())
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			inventoryId := uuid.New()
			var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
			inventoryIdString := inventoryId.String()

			Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

			events := buildSatEvents(
				data.CorrelationID,
				satPlaybookRunUpdateEvent(1, inventoryIdString, ""),
				satPlaybookRunFinishedEvent(inventoryIdString, "failure"),
			)

			code := 1
			errorDescription := "Capsule is down"

			(*events)[1].SatelliteInfrastructureCode = &code
			(*events)[1].SatelliteInfrastructureError = &errorDescription

			instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

			run := fetchRun(data.ID)
			seq := 1
			Expect(run.Status).To(Equal("failure"))
			checkHost(data.ID, "failure", &seq, errorDescription, &inventoryId)
		})

		Describe("response_full false", func() {
			It("infers the logs properly (1 host)", func() {
				var data = test.NewRun(accountNumber())
				data.ResponseFull = true

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())

				inventoryId := uuid.New()
				var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
				inventoryIdString := inventoryId.String()

				Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

				events := buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(0, inventoryIdString, "first console log\n"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(0, inventoryIdString, "first console log\n"),
					satPlaybookRunUpdateEvent(1, inventoryIdString, "second console log"),
					satPlaybookRunFinishedEvent(inventoryIdString, "success"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				run := fetchRun(data.ID)

				Expect(run.Status).To(Equal("success"))
				checkHost(data.ID, "success", utils.IntRef(1), "first console log\nsecond console log", &inventoryId)
			})
		})

		Describe("response_full false", func() {
			It("concatenates the logs for the same host", func() {
				var data = test.NewRun(accountNumber())
				data.ResponseFull = false

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())

				inventoryId := uuid.New()
				var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
				inventoryIdString := inventoryId.String()

				Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

				events := buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(0, inventoryIdString, "first console log\n"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(1, inventoryIdString, "second console log"),
					satPlaybookRunFinishedEvent(inventoryIdString, "success"),
					satPlaybookRunCompletedEvent("success"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				run := fetchRun(data.ID)

				Expect(run.Status).To(Equal("success"))
				checkHost(data.ID, "success", utils.IntRef(1), "first console log\nsecond console log", &inventoryId)
			})

			It("adds indicator in logs for missed host sequence", func() {
				var data = test.NewRun(accountNumber())
				data.ResponseFull = false

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())

				inventoryId := uuid.New()
				var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
				inventoryIdString := inventoryId.String()

				Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

				events := buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(0, inventoryIdString, "first console log\n"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(6, inventoryIdString, "second console log"),
					satPlaybookRunFinishedEvent(inventoryIdString, "success"),
					satPlaybookRunCompletedEvent("success"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				run := fetchRun(data.ID)

				Expect(run.Status).To(Equal("success"))
				checkHost(data.ID, "success", utils.IntRef(6), "first console log\n\\n\\u2026\\nsecond console log", &inventoryId)
			})

			It("event ignored if received out of order", func() {
				var data = test.NewRun(accountNumber())
				data.ResponseFull = false

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())

				inventoryId := uuid.New()
				var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
				inventoryIdString := inventoryId.String()

				Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

				events := buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(1, inventoryIdString, "second console log\n"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(0, inventoryIdString, "first console log\n"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(2, inventoryIdString, "third console log"),
					satPlaybookRunFinishedEvent(inventoryIdString, "success"),
					satPlaybookRunCompletedEvent("success"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				run := fetchRun(data.ID)

				Expect(run.Status).To(Equal("success"))
				checkHost(data.ID, "success", utils.IntRef(2), "\\n\\u2026\\nsecond console log\nthird console log", &inventoryId)
			})

			It("failed status not overridden by out-of-order event", func() {
				var data = test.NewRun(accountNumber())
				data.ResponseFull = false

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())

				inventoryId := uuid.New()
				var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
				inventoryIdString := inventoryId.String()

				Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

				events := buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(1, inventoryIdString, "second console log"),
					satPlaybookRunFinishedEvent(inventoryIdString, "failure"),
					satPlaybookRunCompletedEvent("failure"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(0, inventoryIdString, "first console log\n"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				run := fetchRun(data.ID)

				Expect(run.Status).To(Equal("failure"))
			})

			It("correctly determines run/host status if the finished/completed events are sent separately", func() {
				var data = test.NewRun(accountNumber())
				data.ResponseFull = false

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())

				inventoryId := uuid.New()
				var hostData = test.NewRunHost(data.ID, "running", &inventoryId)
				inventoryIdString := inventoryId.String()

				Expect(db().Create(&hostData).Error).ToNot(HaveOccurred())

				events := buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(0, inventoryIdString, "first console log\n"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunUpdateEvent(1, inventoryIdString, "second console log"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				events = buildSatEvents(
					data.CorrelationID,
					satPlaybookRunFinishedEvent(inventoryIdString, "success"),
					satPlaybookRunCompletedEvent("success"),
				)

				instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

				run := fetchRun(data.ID)

				Expect(run.Status).To(Equal("success"))
				checkHost(data.ID, "success", utils.IntRef(1), "first console log\nsecond console log", &inventoryId)
			})

			Context("with multiple hosts", func() {
				It("infers the run status properly", func() {
					var data = test.NewRun(accountNumber())
					data.ResponseFull = false

					Expect(db().Create(&data).Error).ToNot(HaveOccurred())

					inventoryId1 := uuid.New()
					var host1 = test.NewRunHost(data.ID, "running", &inventoryId1)
					inventoryId1String := inventoryId1.String()

					inventoryId2 := uuid.New()
					var host2 = test.NewRunHost(data.ID, "running", &inventoryId2)
					host2.Host = "localhost2"
					inventoryId2String := inventoryId2.String()

					inventoryId3 := uuid.New()
					var host3 = test.NewRunHost(data.ID, "running", &inventoryId3)
					host3.Host = "localhost3"
					inventoryId3String := inventoryId3.String()

					Expect(db().Create(&host1).Error).ToNot(HaveOccurred())
					Expect(db().Create(&host2).Error).ToNot(HaveOccurred())
					Expect(db().Create(&host3).Error).ToNot(HaveOccurred())

					events := buildSatEvents(
						data.CorrelationID,
						satPlaybookRunUpdateEvent(1, inventoryId1String, ""),
						satPlaybookRunFinishedEvent("fa02c492-587c-4baa-b5e1-e5e9b6083cd1", "success"),
					)

					instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

					run := fetchRun(data.ID)
					Expect(run.Status).To(Equal("running")) // one host complete, the rest still running

					events = buildSatEvents(
						data.CorrelationID,
						satPlaybookRunUpdateEvent(1, inventoryId2String, ""),
						satPlaybookRunFinishedEvent("fa02c492-587c-4baa-b5e1-e5e9b6083cd1", "failure"),
					)

					instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

					run = fetchRun(data.ID)
					Expect(run.Status).To(Equal("running")) // still waiting for the last one

					events = buildSatEvents(
						data.CorrelationID,
						satPlaybookRunUpdateEvent(1, inventoryId3String, ""),
						satPlaybookRunFinishedEvent("fa02c492-587c-4baa-b5e1-e5e9b6083cd1", "success"),
						satPlaybookRunCompletedEvent("failure"),
					)

					instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

					run = fetchRun(data.ID)
					Expect(run.Status).To(Equal("failure"))
				})

				It("logs do not interfere between hosts", func() {
					var data = test.NewRun(accountNumber())
					data.ResponseFull = false

					Expect(db().Create(&data).Error).ToNot(HaveOccurred())

					inventoryId1 := uuid.New()
					var host1 = test.NewRunHost(data.ID, "running", &inventoryId1)
					inventoryId1String := inventoryId1.String()

					inventoryId2 := uuid.New()
					var host2 = test.NewRunHost(data.ID, "running", &inventoryId2)
					host2.Host = "localhost2"
					inventoryId2String := inventoryId2.String()

					Expect(db().Create(&host1).Error).ToNot(HaveOccurred())
					Expect(db().Create(&host2).Error).ToNot(HaveOccurred())

					events := buildSatEvents(
						data.CorrelationID,
						satPlaybookRunUpdateEvent(0, inventoryId1String, "a"),
						satPlaybookRunUpdateEvent(0, inventoryId2String, "f"),
					)

					instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

					events = buildSatEvents(
						data.CorrelationID,
						satPlaybookRunUpdateEvent(1, inventoryId1String, "b"),
					)

					instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

					events = buildSatEvents(
						data.CorrelationID,
						satPlaybookRunUpdateEvent(2, inventoryId1String, "c"),
						satPlaybookRunUpdateEvent(2, inventoryId2String, "h"),
					)

					instance.onMessage(test.TestContext(), newSatResponseMessage(events, data.CorrelationID))

					hosts := fetchHosts(data.ID)
					Expect(hosts).To(HaveLen(2))
					sort.SliceStable(hosts, func(i, j int) bool {
						return hosts[i].Host < hosts[j].Host
					})

					Expect(hosts[0].InventoryID).To(Equal(&inventoryId1))
					Expect(hosts[1].InventoryID).To(Equal(&inventoryId2))
					Expect(hosts[0].Log).To(Equal("abc"))
					Expect(hosts[1].Log).To(Equal("f\\n\\u2026\\nh"))
				})
			})
		})
	})
})
