package message

import (
	"fmt"

	"github.com/google/uuid"
)

const EventExecutorOnStart = "executor_on_start"

type ValidatedMessages struct {
	PlaybookType string
	Playbook     []PlaybookRunResponseMessageYamlEventsElem
	PlaybookSat  []PlaybookSatRunResponseMessageYamlEventsElem
}

func GetCorrelationId(events ValidatedMessages, playbookSatType string) (result uuid.UUID, err error) {
	if events.PlaybookType == playbookSatType {
		result, err = uuid.Parse(events.PlaybookSat[0].CorrelationId)
		return
	}

	for _, event := range events.Playbook {
		if event.Event == EventExecutorOnStart && event.EventData != nil && event.EventData.CrcDispatcherCorrelationId != nil {
			result, err = uuid.Parse(*event.EventData.CrcDispatcherCorrelationId)
			return
		}
	}

	err = fmt.Errorf("Correlation id not found")
	return
}
