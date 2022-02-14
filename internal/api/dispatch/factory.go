package dispatch

import (
	"playbook-dispatcher/internal/api/connectors"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

func NewDispatchManager(config *viper.Viper, cloudConnector connectors.CloudConnectorClient, rateLimiter *rate.Limiter, db *gorm.DB) DispatchManager {
	return &dispatchManager{
		config:         config,
		cloudConnector: cloudConnector,
		db:             db,
		rateLimiter:    rateLimiter,
	}
}
