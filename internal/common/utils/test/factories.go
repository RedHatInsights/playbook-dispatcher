package test

import (
	dbModel "playbook-dispatcher/internal/common/model/db"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
)

func EventExecutorOnStart(correlationId string) messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return messageModel.PlaybookRunResponseMessageYamlEventsElem{
		Event: "executor_on_start",
		EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
			CrcDispatcherCorrelationId: &correlationId,
		},
	}
}

func EventPlaybookOnStart() messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return messageModel.PlaybookRunResponseMessageYamlEventsElem{
		Event:     "playbook_on_start",
		EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{},
	}
}

func EventPlaybookOnPlayStart() messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return messageModel.PlaybookRunResponseMessageYamlEventsElem{
		Event:     "playbook_on_play_start",
		EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{},
	}
}

func EventPlaybookOnTaskStart() messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return messageModel.PlaybookRunResponseMessageYamlEventsElem{
		Event:     "playbook_on_task_start",
		EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{},
	}
}

func EventRunnerOnStart(host string) messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return messageModel.PlaybookRunResponseMessageYamlEventsElem{
		Event: "runner_on_start",
		EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
			Host: &host,
		},
	}
}

func EventRunnerOnOk(host string) messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return messageModel.PlaybookRunResponseMessageYamlEventsElem{
		Event: "runner_on_ok",
		EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
			Host: &host,
		},
	}
}

func EventPlaybookOnStats(host string) messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return messageModel.PlaybookRunResponseMessageYamlEventsElem{
		Event: "playbook_on_stats",
		EventData: &messageModel.PlaybookRunResponseMessageYamlEventsElemEventData{
			Host: &host,
		},
	}
}

func EventSequenceOk(correlationId, host string) []messageModel.PlaybookRunResponseMessageYamlEventsElem {
	return []messageModel.PlaybookRunResponseMessageYamlEventsElem{
		EventExecutorOnStart(correlationId),
		EventPlaybookOnStart(),
		EventPlaybookOnPlayStart(),
		EventPlaybookOnTaskStart(),
		EventRunnerOnStart(host),
		EventRunnerOnOk(host),
		EventPlaybookOnStats(host),
	}
}

func NewRunWithStatus(account string, status string) dbModel.Run {
	return dbModel.Run{
		ID:            uuid.New(),
		Account:       account,
		Recipient:     uuid.New(),
		CorrelationID: uuid.New(),
		URL:           "http://example.com",
		Status:        status,
		Timeout:       3600,
		Service:       "test",
	}
}

func NewRun(account string) dbModel.Run {
	return NewRunWithStatus(account, "running")
}

func NewRunsWithLocalhost(accountNumber string, n int) *[]dbModel.Run {
	runs := make([]dbModel.Run, n)

	for i := 0; i < n; i++ {
		run := NewRun(accountNumber)
		run.Events = utils.MustMarshal(EventSequenceOk(uuid.New().String(), "localhost"))
		runs[i] = run
	}

	return &runs
}
