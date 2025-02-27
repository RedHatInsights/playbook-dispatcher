package sources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"playbook-dispatcher/internal/common/constants"
	"playbook-dispatcher/internal/common/utils"
	"strings"
	"time"

	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/spf13/viper"
)

const (
	basePath               = "/api/sources/v3.1/"
	filterPath QueryFilter = "filter[source_ref][eq]="
)

type sourcesClientImpl struct {
	client ClientWithResponsesInterface
}

func NewSourcesClientWithHttpRequestDoer(cfg *viper.Viper, doer HttpRequestDoer) SourcesConnector {
	client := &ClientWithResponses{
		ClientInterface: &Client{
			Server: fmt.Sprintf("%s://%s:%d%s", cfg.GetString("sources.scheme"), cfg.GetString("sources.host"), cfg.GetInt("sources.port"), basePath),
			Client: utils.NewMeasuredHttpRequestDoer(doer, "sources", "postMessage"),
			RequestEditors: []RequestEditorFn{func(ctx context.Context, req *http.Request) error {
				req.Header.Set(constants.HeaderRequestId, request_id.GetReqID(ctx))

				if identity, ok := ctx.Value(constants.HeaderIdentity).(string); ok {
					req.Header.Set(constants.HeaderIdentity, identity)
				}

				originalUrl := req.URL.String()
				if strings.Contains(originalUrl, "filter=filter%5B") {
					// Remove the extra filter parameter name
					urlWithExtraFilterRemoved := strings.Replace(req.URL.String(), "filter=", "", -1)

					// Convert the url encoded "=" to unencoded "=" for...otherwise sources returns a 400
					urlWithEqualUnencoded := strings.Replace(urlWithExtraFilterRemoved, "%3D", "=", -1)

					newUrl, err := url.Parse(urlWithEqualUnencoded)
					if err != nil {
						return err
					}

					req.URL = newUrl
				}

				return nil
			}},
		},
	}

	return &sourcesClientImpl{
		client: client,
	}
}

func NewSourcesClient(cfg *viper.Viper) SourcesConnector {
	doer := http.Client{
		Timeout: time.Duration(cfg.GetInt64("sources.timeout") * int64(time.Second)),
	}

	return NewSourcesClientWithHttpRequestDoer(cfg, &doer)
}

func (this *sourcesClientImpl) getRHCConnectionStatus(ctx context.Context, sourceId string) (*string, *string, error) {
	utils.GetLogFromContext(ctx).Debugw("Sending Sources RHC Connection Request")

	ID := ID(sourceId)

	params := GetSourcesRhcConnectionParams{}

	res, err := this.client.GetSourcesRhcConnectionWithResponse(ctx, ID, &params)
	if err != nil {
		return nil, nil, err
	}

	if res.HTTPResponse.StatusCode == 404 {
		return nil, nil, fmt.Errorf("RHCStatus Not Found")
	}

	if res.HTTPResponse.StatusCode == 400 {
		return nil, nil, fmt.Errorf("RHCStatus Bad Request")
	}

	if res.JSON200 == nil {
		return nil, nil, fmt.Errorf(`GetRhcConnectionStatus unexpected status code "%d" or content type "%s"`, res.HTTPResponse.StatusCode, res.HTTPResponse.Header.Get("content-type"))
	}

	if res.JSON200.Data == nil || len(*res.JSON200.Data) == 0 {
		return nil, nil, fmt.Errorf("GetRHCConnectionStatus returned an empty response")
	}

	if (*res.JSON200.Data)[0].AvailabilityStatus == nil {
		return nil, nil, fmt.Errorf("GetRHCConnectionStatus returned an empty AvailabilityStatus")
	}

	rhcConnectionAvailabilityStatus := (*res.JSON200.Data)[0].AvailabilityStatus
	rhcConnectionAvailabilityStatusString := string(*rhcConnectionAvailabilityStatus)

	return (*res.JSON200.Data)[0].RhcId, &rhcConnectionAvailabilityStatusString, err
}

func (this *sourcesClientImpl) getSourceIdBySatelliteId(ctx context.Context, satelliteId string) (sourceId string, sourceName string, err error) {
	utils.GetLogFromContext(ctx).Debugw("Sending Sources Request")

	ID := ID(satelliteId)
	queryFilter := filterPath + QueryFilter(ID)

	params := &ListSourcesParams{
		Filter: &queryFilter,
	}

	res, err := this.client.ListSourcesWithResponse(ctx, params)
	if err != nil {
		return "", "", err
	}

	if res.JSON400 != nil {
		return "", "", fmt.Errorf("Source Bad Request")
	}

	if res.JSON200 == nil {
		return "", "", fmt.Errorf(`GetSources unexpected status code "%d" or content type "%s"`, res.HTTPResponse.StatusCode, res.HTTPResponse.Header.Get("content-type"))
	}

	if res.JSON200.Data == nil || len(*res.JSON200.Data) == 0 {
		return "", "", fmt.Errorf("GetSources returned an empty response")
	}

	source := (*res.JSON200.Data)[0]

	if source.Id == nil {
		return "", "", fmt.Errorf("GetSources did not return a valid sources id")
	}

	return string(*source.Id), *source.Name, nil
}

func (this *sourcesClientImpl) GetSourceConnectionDetails(ctx context.Context, sourceID string) (details SourceConnectionStatus, err error) {
	utils.GetLogFromContext(ctx).Debugw("Gathering Source Connection Details")

	sourceId, sourceName, err := this.getSourceIdBySatelliteId(ctx, sourceID)
	if err != nil {
		return SourceConnectionStatus{}, err
	}

	rhcId, availabilityStatus, err := this.getRHCConnectionStatus(ctx, sourceId)
	if err != nil {
		return SourceConnectionStatus{}, err
	}

	return SourceConnectionStatus{
		ID:                 sourceId,
		SourceName:         &sourceName,
		RhcID:              rhcId,
		AvailabilityStatus: availabilityStatus,
	}, err
}
