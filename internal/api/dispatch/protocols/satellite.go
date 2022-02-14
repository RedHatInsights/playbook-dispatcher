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

func getHostsLine(runInput generic.RunInput) string {
	hosts := make([]string, len(runInput.Hosts))
	for i, host := range runInput.Hosts {
		hosts[i] = (*host.InventoryId).String()
	}

	return strings.Join(hosts, ",")
}

func (this *satelliteProtocol) BuildMedatada(runInput generic.RunInput, correlationID uuid.UUID, cfg *viper.Viper) map[string]string {
	hosts := getHostsLine(runInput)

	principalHash := fmt.Sprintf("%x", sha256.Sum256([]byte(*runInput.Principal)))

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
