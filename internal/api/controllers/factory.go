package controllers

import (
	"playbook-dispatcher/internal/api/connectors"

	"gorm.io/gorm"
)

func CreateControllers(database *gorm.DB, cloudConnectorClient connectors.CloudConnectorClient) ServerInterface {
	return &controllers{
		database:             database,
		cloudConnectorClient: cloudConnectorClient,
	}
}

// implements api.ServerInterface
type controllers struct {
	database             *gorm.DB
	cloudConnectorClient connectors.CloudConnectorClient
}
