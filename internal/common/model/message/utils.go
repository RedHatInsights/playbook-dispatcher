package message

import (
	"errors"

	"github.com/google/uuid"
)

// ErrCorrelationIDNotFound is returned when a validated payload has no
// dispatcher correlation id. Invalid id strings return the uuid.Parse error.
var ErrCorrelationIDNotFound = errors.New("correlation id not found")

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
		if event.EventData != nil && event.EventData.CrcDispatcherCorrelationId != nil {
			result, err = uuid.Parse(*event.EventData.CrcDispatcherCorrelationId)
			return
		}
	}

	err = ErrCorrelationIDNotFound
	return
}
