package private

import (
	"fmt"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/common/config"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
	"gorm.io/gorm"
)

func CreateController(database *gorm.DB, cloudConnectorClient connectors.CloudConnectorClient, config *viper.Viper) ServerInterfaceWrapper {
	return ServerInterfaceWrapper{
		Handler: &controllers{
			database:             database,
			cloudConnectorClient: cloudConnectorClient,
			config:               config,
			rateLimiter:          ratelimit.New(config.GetInt("cloud.connector.rps")),
		},
	}
}

// implements api.ServerInterface
type controllers struct {
	database             *gorm.DB
	cloudConnectorClient connectors.CloudConnectorClient
	config               *viper.Viper
	rateLimiter          ratelimit.Limiter
}

// workaround for https://github.com/deepmap/oapi-codegen/issues/42
func GetSwaggerWithExternalRefs() (*openapi3.Swagger, error) {
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true

	swagger, err := loader.LoadSwaggerFromFile(config.Get().GetString("schema.api.private"))
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}

	return swagger, nil
}
