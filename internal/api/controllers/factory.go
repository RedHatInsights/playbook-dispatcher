package controllers

import (
	"playbook-dispatcher/internal/api/connectors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateControllers(database *gorm.DB, log *zap.SugaredLogger, cloudConnectorClient connectors.CloudConnectorClient) ServerInterface {
	return &controllers{
		database:             database,
		log:                  log,
		cloudConnectorClient: cloudConnectorClient,
	}
}

// implements api.ServerInterface
type controllers struct {
	database             *gorm.DB
	log                  *zap.SugaredLogger
	cloudConnectorClient connectors.CloudConnectorClient
}
