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

func (this *satelliteProtocol) GetDirective() Directive {
	return SatelliteDirective
}

func (this *satelliteProtocol) GetLabel() string {
	return LabelSatRequest
}

func (this *satelliteProtocol) GetResponseFull(cfg *viper.Viper) bool {
	return cfg.GetBool("satellite.response.full")
}

func (this *satelliteProtocol) GetPrincipalHash(principal string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(principal)))
}

func getHostsLine(runInput generic.RunInput) string {
	hosts := make([]string, len(runInput.Hosts))
	for i, host := range runInput.Hosts {
		hosts[i] = (*host.InventoryId).String()
	}

	return strings.Join(hosts, ",")
}

func (this *satelliteProtocol) BuildMedatada(runInput generic.RunInput, correlationID uuid.UUID, cfg *viper.Viper) map[string]string {
	hosts := getHostsLine(runInput)

	principalHash := this.GetPrincipalHash(*runInput.Principal)

	metadata := buildCommonSignal(cfg)

	metadata["operation"] = "run"
	metadata["correlation_id"] = correlationID.String()
	metadata["playbook_run_name"] = *runInput.Name
	metadata["playbook_run_url"] = string(*runInput.WebConsoleUrl)
	metadata["sat_id"] = (*runInput.SatId).String()
	metadata["sat_org_id"] = *runInput.SatOrgId
	metadata["initiator_user_id"] = principalHash
	metadata["hosts"] = hosts
	metadata["response_full"] = strconv.FormatBool(this.GetResponseFull(cfg))

	return metadata
}

func (this *satelliteProtocol) BuildCancelMetadata(cancelInput generic.CancelInput, correlationID uuid.UUID, cfg *viper.Viper) map[string]string {
	principalHash := this.GetPrincipalHash(cancelInput.Principal)

	metadata := map[string]string{
		"operation":         "cancel",
		"correlation_id":    correlationID.String(),
		"initiator_user_id": principalHash,
	}

	return metadata
}
