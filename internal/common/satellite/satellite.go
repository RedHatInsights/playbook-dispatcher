package satellite

import (
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"sort"
)

func GetSatHosts(events []messageModel.PlaybookSatRunResponseMessageYamlEventsElem) []string {
	hosts := make(map[string]interface{})

	for _, event := range events {
		if event.Host != nil {
			hosts[*event.Host] = true
		}
	}

	keys := utils.MapKeys(hosts)
	sort.Strings(keys)
	return keys
}

type SatHostInfo struct {
	Sequence int
	Console  string
}

func GetSatHostInfo(events []messageModel.PlaybookSatRunResponseMessageYamlEventsElem, host *string) *SatHostInfo {
	hostInfo := SatHostInfo{}
	for _, event := range events {
		if event.Host != nil && *event.Host != *host {
			continue
		}
		if event.Sequence != nil {
			hostInfo.Sequence = *event.Sequence
		}
		if event.Console != nil {
			hostInfo.Console += *event.Console
		}

		if event.SatelliteConnectionCode != nil && *event.SatelliteConnectionCode != 0 {
			hostInfo.Console += *event.SatelliteConnectionError
		}
		if event.SatelliteInfrastructureCode != nil && *event.SatelliteInfrastructureCode != 0 {
			hostInfo.Console += *event.SatelliteInfrastructureError
		}
	}
	return &hostInfo
}
