package inventory

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/constants"
	"playbook-dispatcher/internal/common/utils"
	"time"

	"github.com/google/uuid"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
	"github.com/spf13/viper"
)

const basePath = "/api/inventory/v1/hosts"

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
				strigifiedInstanceID := convertInterfaceToString(satelliteInstanceID)
				satelliteFacts.SatelliteInstanceID = &strigifiedInstanceID
			}

			if versionExists {
				stringifiedVersion := convertInterfaceToString(satelliteVersion)
				satelliteFacts.SatelliteVersion = &stringifiedVersion
			}

			if orgIDExists {
				stringifiedOrgID := convertInterfaceToString(satelliteOrgID)
				satelliteFacts.SatelliteOrgID = &stringifiedOrgID
			}
		}
	}

	return satelliteFacts
}

func convertInterfaceToString(v interface{}) string {
	switch result := v.(type) {
	case float64:
		return fmt.Sprintf("%d", int(result))
	case string:
		return v.(string)
	default:
		return ""
	}
}

func createHostGetHostByIdParams(orderBy string, orderHow string) *ApiHostGetHostByIdParams {
	orderByParam := ApiHostGetHostByIdParamsOrderBy(orderBy)
	orderHowParam := ApiHostGetHostByIdParamsOrderHow(orderHow)

	return &ApiHostGetHostByIdParams{
		OrderBy:  &orderByParam,
		OrderHow: &orderHowParam,
	}
}

func createHostGetHostSystemProfileByIdParams(orderBy string, orderHow string) *ApiHostGetHostSystemProfileByIdParams {
	orderByParam := ApiHostGetHostSystemProfileByIdParamsOrderBy(orderBy)
	orderHowParam := ApiHostGetHostSystemProfileByIdParamsOrderHow(orderHow)

	fields := FieldsParam(
		SystemProfileNestedObject{
			"fields[system_profile]": SystemProfileNestedObject_AdditionalProperties{[]byte(`["rhc_client_id", "owner_id"]`)},
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
			RequestEditors: []RequestEditorFn{func(ctx context.Context, req *http.Request) error {
				req.Header.Set(constants.HeaderRequestId, request_id.GetReqID(ctx))

				if identity, ok := ctx.Value(constants.HeaderIdentity).(string); ok {
					req.Header.Set(constants.HeaderIdentity, identity)
				}

				return nil
			}},
		},
	}

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

	clientIds, err := strSliceToUUIDSlice(IDs)
	if err != nil {
		return nil, err
	}

	params := createHostGetHostByIdParams(orderBy, orderHow)

	response, err := this.client.ApiHostGetHostByIdWithResponse(ctx, clientIds, params)

	if err != nil {
		return nil, err
	}

	if response.StatusCode() == http.StatusNotFound {
		return []HostOut{}, nil
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

	clientIds, err := strSliceToUUIDSlice(IDs)
	if err != nil {
		return nil, err
	}

	params := createHostGetHostSystemProfileByIdParams(orderBy, orderHow)

	response, err := this.client.ApiHostGetHostSystemProfileByIdWithResponse(ctx, clientIds, params)

	if err != nil {
		return nil, err
	}

	if response.JSON200 == nil {
		return nil, utils.UnexpectedResponse(response.HTTPResponse)
	}

	formatedResults := keySystemProfileResults(response.JSON200.Results)

	return formatedResults, nil
}

func (this *inventoryConnectorImpl) GetHostConnectionDetails(ctx context.Context, IDs []string, order_by string, order_how string, limit int, offset int) (details []HostDetails, err error) {

	hostResults, err := this.getHostDetails(ctx, IDs, order_by, order_how, limit, offset)

	if err != nil {
		return nil, err
	}

	if len(hostResults) == 0 {
		return []HostDetails{}, nil
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

func strSliceToUUIDSlice(strSlice []string) ([]uuid.UUID, error) {
	uuidSlice := make([]uuid.UUID, 0, len(strSlice))

	for _, str := range strSlice {
		uuid, err := uuid.Parse(str)
		if err != nil {
			return nil, err
		}

		uuidSlice = append(uuidSlice, uuid)
	}

	return uuidSlice, nil
}
