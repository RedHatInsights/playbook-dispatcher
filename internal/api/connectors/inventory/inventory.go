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

const basePath = "/api/inventory/"

type InventoryConnector interface {
	GetHostDetails(
		ctx context.Context,
		IDs []string,
		orderBy string,
		orderHow string,
		limit int,
		offest int,
	) (hostDetails []HostDetails, err error)

	GetSystemProfileDetails(
		ctx context.Context,
		IDs []string,
		orderBy string,
		orderHow string,
		limit int,
		offset int,
	) (profileDetails []SystemProfileDetails, err error)
}

type inventoryConnectorImpl struct {
	client ClientWithResponsesInterface
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

func createHostDetails(response *ApiHostGetHostByIdResponse) []HostDetails {
	hostDetails := make([]HostDetails, len(response.JSON200.Results))

	for i, response := range response.JSON200.Results {
		satelliteFacts := getSatelliteFacts(response.Facts)

		hostDetails[i] = HostDetails{
			ID:          *response.Id,
			DisplayName: *response.DisplayName,
			Facts:       satelliteFacts,
			CanonicalFacts: map[string]interface{}{
				"fqdn": *response.CanonicalFactsOut.Fqdn,
			},
		}
	}

	return hostDetails
}

func createSystemProfileDetails(response *ApiHostGetHostSystemProfileByIdResponse) []SystemProfileDetails {
	systemProfileDetails := make([]SystemProfileDetails, len(response.JSON200.Results))
	for i, response := range response.JSON200.Results {
		systemProfileDetails[i] = SystemProfileDetails{
			ID: *response.Id,
			SystemProfileFacts: map[string]interface{}{
				"rhc_client_id": *response.SystemProfile.RhcClientId,
				"owner_id":      *response.SystemProfile.OwnerId,
			},
		}
	}

	return systemProfileDetails
}

func getSatelliteFacts(facts *[]FactSet) map[string]interface{} {
	satelliteFacts := map[string]interface{}{}
	for _, fact := range *facts {
		if fact.Namespace == "satellite" {
			satelliteFacts = fact.Facts
		}
	}

	return satelliteFacts
}

func NewInventoryClientWithHttpRequestDoer(cfg *viper.Viper, doer HttpRequestDoer) InventoryConnector {
	client := &ClientWithResponses{
		ClientInterface: &Client{
			Server: fmt.Sprintf("%s://%s:%d%s", cfg.GetString("inventory.connector.scheme"), cfg.GetString("inventory.connector.host"), cfg.GetInt("inventory.connector.port"), basePath),
			Client: utils.NewMeasuredHttpRequestDoer(doer, "inventory", "GetHostDetails"), // TODO Change this
			RequestEditor: func(ctx context.Context, req *http.Request) error {
				req.Header.Set(constants.HeaderRequestId, request_id.GetReqID(ctx))

				return nil
			},
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

func (this *inventoryConnectorImpl) GetHostDetails(
	ctx context.Context,
	IDs []string,
	orderBy string,
	orderHow string,
	limit int,
	offset int,
) (details []HostDetails, err error) {

	params := createHostGetHostByIdParams(orderBy, orderHow)

	response, err := this.client.ApiHostGetHostByIdWithResponse(ctx, IDs, params)

	if err != nil {
		return []HostDetails{}, err
	}

	if response.JSON200 == nil {
		return []HostDetails{}, utils.UnexpectedResponse(response.HTTPResponse)
	}

	hostDetails := createHostDetails(response)

	return hostDetails, err
}

func (this *inventoryConnectorImpl) GetSystemProfileDetails(
	ctx context.Context,
	IDs []string,
	orderBy string,
	orderHow string,
	limit int,
	offset int,
) (details []SystemProfileDetails, err error) {

	params := createHostGetHostSystemProfileByIdParams(orderBy, orderHow)

	response, err := this.client.ApiHostGetHostSystemProfileByIdWithResponse(ctx, IDs, params)

	if err != nil {
		return []SystemProfileDetails{}, err
	}

	if response.JSON200 == nil {
		return []SystemProfileDetails{}, utils.UnexpectedResponse(response.HTTPResponse)
	}

	systemProfileDetails := createSystemProfileDetails(response)

	return systemProfileDetails, nil
}
