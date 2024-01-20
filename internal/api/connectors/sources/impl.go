package sources

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
			RequestEditor: func(ctx context.Context, req *http.Request) error {
				req.Header.Set(constants.HeaderRequestId, request_id.GetReqID(ctx))

				if identity, ok := ctx.Value(constants.HeaderIdentity).(string); ok {
					req.Header.Set(constants.HeaderIdentity, identity)
				}

				originalUrl := req.URL.String()
				if strings.Contains(originalUrl, "filter=filter[") {
					// Remove the extra filter parameter name
					correctedUrl := strings.Replace(req.URL.String(), "filter=", "", -1)

					newUrl, err := url.Parse(correctedUrl)
					if err != nil {
						return err
					}

					req.URL = newUrl
				}

				return nil
			},
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

func (this *sourcesClientImpl) getRHCConnectionStatus(ctx context.Context, sourceId string) (status *RhcConnectionRead, err error) {
	utils.GetLogFromContext(ctx).Debugw("Sending Sources RHC Connection Request")

	ID := ID(sourceId)

	res, err := this.client.GetRhcConnectionWithResponse(ctx, ID)

	if err != nil {
		return nil, err
	}

	if res.HTTPResponse.StatusCode == 404 {
		return nil, fmt.Errorf("RHCStatus Not Found")
	}

	if res.HTTPResponse.StatusCode == 400 {
		return nil, fmt.Errorf("RHCStatus Bad Request")
	}

	if res.JSON200 == nil {
		return nil, fmt.Errorf(`GetRhcConnectionStatus unexpected status code "%d" or content type "%s"`, res.HTTPResponse.StatusCode, res.HTTPResponse.Header.Get("content-type"))
	}

	return res.JSON200, err
}

func (this *sourcesClientImpl) getSources(ctx context.Context, sourceId string) (sources *[]Source, err error) {
	utils.GetLogFromContext(ctx).Debugw("Sending Sources Request")

	ID := ID(sourceId)
	queryFilter := filterPath + QueryFilter(ID)

	params := &ListSourcesParams{
		Filter: &queryFilter,
	}

	res, err := this.client.ListSourcesWithResponse(ctx, params)

	if err != nil {
		return nil, err
	}

	if res.JSON400 != nil {
		return nil, fmt.Errorf("Source Bad Request")
	}

	if res.JSON200 == nil {
		return nil, fmt.Errorf(`GetSources unexpected status code "%d" or content type "%s"`, res.HTTPResponse.StatusCode, res.HTTPResponse.Header.Get("content-type"))
	}

	return res.JSON200.Data, err
}

func (this *sourcesClientImpl) GetSourceConnectionDetails(ctx context.Context, sourceID string) (details SourceConnectionStatus, err error) {
	utils.GetLogFromContext(ctx).Debugw("Gathering Source Connection Details")

	sourcesResponse, err := this.getSources(ctx, sourceID)

	if err != nil {
		return SourceConnectionStatus{}, err
	}

	rhcConnectionResponse, err := this.getRHCConnectionStatus(ctx, sourceID)

	if err != nil {
		return SourceConnectionStatus{}, err
	}

	source := (*sourcesResponse)[0]
	return SourceConnectionStatus{
		ID:                 string(*source.Id),
		SourceName:         source.Name,
		RhcID:              rhcConnectionResponse.RhcId,
		AvailabilityStatus: rhcConnectionResponse.AvailabilityStatus,
	}, err
}
