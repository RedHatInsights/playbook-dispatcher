package protocols

import (
	"playbook-dispatcher/internal/common/model/generic"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type runnerProtocol struct{}

func (rp *runnerProtocol) GetDirective() Directive {
	return RunnerDirective
}

func (rp *runnerProtocol) GetLabel() string {
	return LabelRunnerRequest
}

func (rp *runnerProtocol) GetResponseFull(cfg *viper.Viper) bool {
	return true
}

func (rp *runnerProtocol) BuildMetaData(runInput generic.RunInput, correlationID uuid.UUID, cfg *viper.Viper) map[string]string {
	metadata := buildCommonSignal(cfg)
	metadata["crc_dispatcher_correlation_id"] = correlationID.String()

	return metadata
}
