package private

import (
	"fmt"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/common/config"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

func CreateController(database *gorm.DB, cloudConnectorClient connectors.CloudConnectorClient, config *viper.Viper) ServerInterfaceWrapper {
	return ServerInterfaceWrapper{
		Handler: &controllers{
			database:             database,
			cloudConnectorClient: cloudConnectorClient,
			config:               config,
			rateLimiter:          getRateLimiter(config),
		},
	}
}

// implements api.ServerInterface
type controllers struct {
	database             *gorm.DB
	cloudConnectorClient connectors.CloudConnectorClient
	config               *viper.Viper
	rateLimiter          *rate.Limiter
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

// returns a rate limiter reference that uses the token-bucket algorithm
func getRateLimiter(config *viper.Viper) *rate.Limiter {
	limit := rate.Limit(config.GetInt("cloud.connector.rps"))
	bucket := config.GetInt("cloud.connector.req.bucket")
	return rate.NewLimiter(limit, bucket)
}
