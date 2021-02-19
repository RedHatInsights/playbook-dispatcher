package public

import (
	"playbook-dispatcher/internal/api/connectors"

	"gorm.io/gorm"
)

func CreateController(database *gorm.DB, cloudConnectorClient connectors.CloudConnectorClient) ServerInterfaceWrapper {
	return ServerInterfaceWrapper{
		Handler: &controllers{
			database:             database,
			cloudConnectorClient: cloudConnectorClient,
		},
	}
}

// implements api.ServerInterface
type controllers struct {
	database             *gorm.DB
	cloudConnectorClient connectors.CloudConnectorClient
}
