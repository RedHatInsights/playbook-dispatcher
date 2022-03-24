package dispatch

import (
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/model/generic"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func newRun(input *generic.RunInput, correlationId uuid.UUID, responseFull bool, service string, cfg *viper.Viper) dbModel.Run {
	run := dbModel.Run{
		ID:             uuid.New(),
		Account:        input.Account,
		CorrelationID:  correlationId,
		URL:            input.Url,
		Status:         dbModel.RunStatusRunning,
		Recipient:      input.Recipient,
		Labels:         input.Labels,
		ResponseFull:   responseFull,
		Service:        service,
		Timeout:        *input.Timeout,       // defaulted
		PlaybookRunUrl: *input.WebConsoleUrl, // defaulted
		PlaybookName:   input.Name,
		Principal:      input.Principal,
		SatId:          input.SatId,
		SatOrgId:       input.SatOrgId,
	}

	if input.OrgId != nil {
		run.OrgID = *input.OrgId
	}

	return run
}

func newHostRun(runHosts []generic.RunHostsInput, entityId uuid.UUID) []dbModel.RunHost {
	newHosts := make([]dbModel.RunHost, len(runHosts))

	for i, inputHost := range runHosts {
		newHosts[i] = dbModel.RunHost{
			ID:          uuid.New(),
			RunID:       entityId,
			InventoryID: inputHost.InventoryId,
			Status:      dbModel.RunStatusRunning,
		}

		if inputHost.AnsibleHost != nil {
			newHosts[i].Host = *inputHost.AnsibleHost
		} else {
			newHosts[i].Host = inputHost.InventoryId.String()
		}
	}

	return newHosts
}
