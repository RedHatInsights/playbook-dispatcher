package inventory

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/constants"
	"playbook-dispatcher/internal/common/utils"
	"time"

	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/spf13/viper"
)

const basePath = "/api/inventory/v1/hosts/"

type inventoryConnectorImpl struct {
	client ClientWithResponsesInterface
}

func keySystemProfileResults(systemProfileResults []HostSystemProfileOut) map[string]HostSystemProfileOut {
	formatedResults := make(map[string]HostSystemProfileOut, len(systemProfileResults))

	for _, result := range systemProfileResults {
		formatedResults[*result.Id] = result
	}

	return formatedResults
}

func getSatelliteFacts(facts *[]FactSet) satelliteFacts {
	satelliteFacts := satelliteFacts{}
	for _, fact := range *facts {
		if fact.Namespace == "satellite" {
			satelliteInstanceID, idExists := fact.Facts["satellite_instance_id"]
			satelliteVersion, versionExists := fact.Facts["satellite_version"]
			satelliteOrgID, orgIDExists := fact.Facts["organization_id"]

			if idExists {
				strigifiedInstanceID := satelliteInstanceID.(string)
				satelliteFacts.SatelliteInstanceID = &strigifiedInstanceID
			}

			if versionExists {
				stringifiedVersion := satelliteVersion.(string)
				satelliteFacts.SatelliteVersion = &stringifiedVersion
			}

			if orgIDExists {
				stringifiedOrgID := satelliteOrgID.(string)
				satelliteFacts.SatelliteOrgID = &stringifiedOrgID
			}
		}
	}

	return satelliteFacts
}

func createHostGetHostByIdParams(orderBy string, orderHow string) *ApiHostGetHostByIdParams {
	orderByParam := OrderByParam(orderBy)
	orderHowParam := OrderHowParam(OrderHowParam_ASC)

	return &ApiHostGetHostByIdParams{
		OrderBy:  &orderByParam,
		OrderHow: &orderHowParam,
	}
}

func createHostGetHostSystemProfileByIdParams(orderBy string, orderHow string) *ApiHostGetHostSystemProfileByIdParams {
	orderByParam := OrderByParam(orderBy)
	orderHowParam := OrderHowParam(orderHow)
	fields := FieldsParam(
		SystemProfileNestedObject{
			AdditionalProperties: map[string]interface{}{
				"fields[system_profile]": []string{"rhc_client_id", "owner_id"},
			},
		},
	)

	return &ApiHostGetHostSystemProfileByIdParams{
		OrderBy:  &orderByParam,
		OrderHow: &orderHowParam,
		Fields:   &fields,
	}
}

func NewInventoryClientWithHttpRequestDoer(cfg *viper.Viper, doer HttpRequestDoer) InventoryConnector {
	client := &ClientWithResponses{
		ClientInterface: &Client{
			Server: fmt.Sprintf("%s://%s:%d%s", cfg.GetString("inventory.connector.scheme"), cfg.GetString("inventory.connector.host"), cfg.GetInt("inventory.connector.port"), basePath),
			Client: utils.NewMeasuredHttpRequestDoer(doer, "inventory", "GetHostConnectionDetails"),
			RequestEditor: func(ctx context.Context, req *http.Request) error {
				req.Header.Set(constants.HeaderRequestId, request_id.GetReqID(ctx))

				if identity, ok := ctx.Value(constants.HeaderIdentity).(string); ok {
					req.Header.Set(constants.HeaderIdentity, identity)
				}

				return nil
			},
		},
	}

    fmt.Println("** inventory client.Server:", client.Server)

	return &inventoryConnectorImpl{
		client: client,
	}
}

func NewInventoryClient(cfg *viper.Viper) InventoryConnector {
	httpClient := http.Client{
		Timeout: time.Duration(cfg.GetInt64("inventory.connector.timeout") * int64(time.Second)),
	}

	return NewInventoryClientWithHttpRequestDoer(cfg, &httpClient)
}

func (this *inventoryConnectorImpl) getHostDetails(
	ctx context.Context,
	IDs []string,
	orderBy string,
	orderHow string,
	limit int,
	offset int,
) (details []HostOut, err error) {

	params := createHostGetHostByIdParams(orderBy, orderHow)

	response, err := this.client.ApiHostGetHostByIdWithResponse(ctx, IDs, params)

	if err != nil {
		return nil, err
	}

	if response.JSON200 == nil {
		return nil, utils.UnexpectedResponse(response.HTTPResponse)
	}

	return response.JSON200.Results, err
}

func (this *inventoryConnectorImpl) getSystemProfileDetails(
	ctx context.Context,
	IDs []string,
	orderBy string,
	orderHow string,
	limit int,
	offset int,
) (details map[string]HostSystemProfileOut, err error) {

	params := createHostGetHostSystemProfileByIdParams(orderBy, orderHow)

	response, err := this.client.ApiHostGetHostSystemProfileByIdWithResponse(ctx, IDs, params)

	if err != nil {
		return nil, err
	}

	if response.JSON200 == nil {
		return nil, utils.UnexpectedResponse(response.HTTPResponse)
	}

	formatedResults := keySystemProfileResults(response.JSON200.Results)

	return formatedResults, nil
}

func (this *inventoryConnectorImpl) GetHostConnectionDetails(ctx context.Context, IDs []string, order_how string, order_by string, limit int, offset int) (details []HostDetails, err error) {

	hostResults, err := this.getHostDetails(ctx, IDs, order_how, order_by, limit, offset)

	if err != nil {
		return nil, err
	}

	systemProfileResults, err := this.getSystemProfileDetails(ctx, IDs, order_by, order_how, limit, offset)

	if err != nil {
		return nil, err
	}

	hostConnectionDetails := make([]HostDetails, len(IDs))
	for i, host := range hostResults {
		satelliteFacts := getSatelliteFacts(host.Facts)
		hostConnectionDetails[i] = HostDetails{
			ID:                  *host.Id,
			OwnerID:             systemProfileResults[*host.Id].SystemProfile.OwnerId,
			SatelliteInstanceID: satelliteFacts.SatelliteInstanceID,
			SatelliteVersion:    satelliteFacts.SatelliteVersion,
			SatelliteOrgID:      satelliteFacts.SatelliteOrgID,
			RHCClientID:         systemProfileResults[*host.Id].SystemProfile.RhcClientId,
		}
	}

	return hostConnectionDetails, nil
}
