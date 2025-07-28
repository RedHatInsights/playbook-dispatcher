package protocols

import (
	"crypto/sha256"
	"fmt"
	"playbook-dispatcher/internal/common/model/generic"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type satelliteProtocol struct{}

func (sp *satelliteProtocol) GetDirective() Directive {
	return SatelliteDirective
}

func (sp *satelliteProtocol) GetLabel() string {
	return LabelSatRequest
}

func (sp *satelliteProtocol) GetResponseFull(cfg *viper.Viper) bool {
	return cfg.GetBool("satellite.response.full")
}

func (sp *satelliteProtocol) GetPrincipalHash(principal string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(principal)))
}

func getHostsLine(runInput generic.RunInput) string {
	hosts := make([]string, len(runInput.Hosts))
	for i, host := range runInput.Hosts {
		hosts[i] = (*host.InventoryId).String()
	}

	return strings.Join(hosts, ",")
}

// collates existing Subscription Manager IDs into a comma-delimited string
func submanIdsAsString(runInput generic.RunInput) string {
	submanIDs := make([]string, 0)
	for _, host := range runInput.Hosts {
		if host.SubscriptionManagerId != nil && host.SubscriptionManagerId.String() != "" {
			submanIDs = append(submanIDs, (*host.SubscriptionManagerId).String())
		}
	}

	if len(submanIDs) > 0 {
		return strings.Join(submanIDs, ",")
	}

	return ""
}

func (sp *satelliteProtocol) BuildMetaData(runInput generic.RunInput, correlationID uuid.UUID, cfg *viper.Viper) map[string]string {
	hosts := getHostsLine(runInput)
	submanIDs := submanIdsAsString(runInput)

	principalHash := sp.GetPrincipalHash(*runInput.Principal)

	metadata := buildCommonSignal(cfg)

	metadata["operation"] = "run"
	metadata["correlation_id"] = correlationID.String()
	metadata["playbook_run_name"] = *runInput.Name
	metadata["playbook_run_url"] = string(*runInput.WebConsoleUrl)
	metadata["sat_id"] = (*runInput.SatId).String()
	metadata["sat_org_id"] = *runInput.SatOrgId
	metadata["initiator_user_id"] = principalHash
	metadata["hosts"] = hosts
	// maintain compatibility if no subscription manager ids are provided
	if submanIDs != "" {
		metadata["subscription_manager_ids"] = submanIDs
	}
	metadata["response_full"] = strconv.FormatBool(sp.GetResponseFull(cfg))

	return metadata
}

func (sp *satelliteProtocol) BuildCancelMetaData(cancelInput generic.CancelInput, correlationID uuid.UUID, cfg *viper.Viper) map[string]string {
	principalHash := sp.GetPrincipalHash(cancelInput.Principal)

	metadata := map[string]string{
		"operation":         "cancel",
		"correlation_id":    correlationID.String(),
		"initiator_user_id": principalHash,
	}

	return metadata
}
