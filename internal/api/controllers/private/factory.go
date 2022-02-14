package private

import (
	"fmt"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/connectors/tenants"
	"playbook-dispatcher/internal/api/dispatch"
	"playbook-dispatcher/internal/common/config"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

func CreateController(database *gorm.DB, cloudConnectorClient connectors.CloudConnectorClient, config *viper.Viper, translator tenants.TenantIDTranslator) ServerInterfaceWrapper {
	rateLimiter := getRateLimiter(config)

	return ServerInterfaceWrapper{
		Handler: &controllers{
			cloudConnectorClient: cloudConnectorClient,
			config:               config,
			rateLimiter:          rateLimiter,
			translator:           translator,
			dispatchManager:      dispatch.NewDispatchManager(config, cloudConnectorClient, rateLimiter, database),
		},
	}
}

// implements api.ServerInterface
type controllers struct {
	cloudConnectorClient connectors.CloudConnectorClient
	config               *viper.Viper
	rateLimiter          *rate.Limiter
	translator           tenants.TenantIDTranslator
	dispatchManager      dispatch.DispatchManager
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
