package ansible

import (
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"sort"
)

func GetAnsibleHosts(events []messageModel.PlaybookRunResponseMessageYamlEventsElem) []string {
	hosts := make(map[string]interface{})

	for _, event := range events {
		if event.EventData != nil && event.EventData.Host != nil {
			hosts[*event.EventData.Host] = true
		}
	}

	keys := utils.MapKeys(hosts)
	sort.Strings(keys)
	return keys
}

func GetStdout(events []messageModel.PlaybookRunResponseMessageYamlEventsElem) (result string) {
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Counter < events[j].Counter
	})

	for _, event := range events {
		if event.Stdout != nil {
			result += *event.Stdout
		}
	}

	return
}
