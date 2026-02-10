package private

import (
	"fmt"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/connectors/inventory"
	"playbook-dispatcher/internal/api/connectors/sources"
	"playbook-dispatcher/internal/api/dispatch"
	"playbook-dispatcher/internal/common/config"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

func CreateController(database *gorm.DB, cloudConnectorClient connectors.CloudConnectorClient, inventoryConnectorClient inventory.InventoryConnector, sourcesConnectorClient sources.SourcesConnector, config *viper.Viper, translator tenantid.Translator) ServerInterfaceWrapper {
	rateLimiter := getRateLimiter(config)

	return ServerInterfaceWrapper{
		Handler: &controllers{
			database:                 database,
			cloudConnectorClient:     cloudConnectorClient,
			inventoryConnectorClient: inventoryConnectorClient,
			sourcesConnectorClient:   sourcesConnectorClient,
			config:                   config,
			rateLimiter:              rateLimiter,
			translator:               translator,
			dispatchManager:          dispatch.NewDispatchManager(config, cloudConnectorClient, rateLimiter, database),
		},
	}
}

// implements api.ServerInterface
type controllers struct {
	database                 *gorm.DB
	cloudConnectorClient     connectors.CloudConnectorClient
	inventoryConnectorClient inventory.InventoryConnector
	sourcesConnectorClient   sources.SourcesConnector
	config                   *viper.Viper
	rateLimiter              *rate.Limiter
	translator               tenantid.Translator
	dispatchManager          dispatch.DispatchManager
}

// workaround for https://github.com/deepmap/oapi-codegen/issues/42
func GetSwaggerWithExternalRefs() (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	swagger, err := loader.LoadFromFile(config.Get().GetString("schema.api.private"))
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
