package message

import (
	"fmt"

	"github.com/google/uuid"
)

const EventExecutorOnStart = "executor_on_start"

func GetCorrelationId(events []PlaybookRunResponseMessageYamlEventsElem) (result uuid.UUID, err error) {
	for _, event := range events {
		if event.Event == EventExecutorOnStart && event.EventData != nil && event.EventData.CrcDispatcherCorrelationId != nil {
			result, err = uuid.Parse(*event.EventData.CrcDispatcherCorrelationId)
			return
		}
	}

	err = fmt.Errorf("Correlation id not found")
	return
}
