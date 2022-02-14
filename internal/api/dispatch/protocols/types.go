package protocols

import (
	"playbook-dispatcher/internal/common/model/generic"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type Directive string

// represents a builder that knows how to format a message for a particular rhc worker
type Protocol interface {

	// the directive that identifies the given rhc worker
	GetDirective() Directive

	// label used in our metrics to distinguish this type of request
	GetLabel() string

	GetResponseFull(cfg *viper.Viper) bool

	// build the metadata dictionary in a format that the given rhc worker understands
	BuildMedatada(runInput generic.RunInput, correlationID uuid.UUID, cfg *viper.Viper) map[string]string
}
