package connectors

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/constants"
	"time"

	"playbook-dispatcher/internal/common/utils"

	"github.com/redhatinsights/platform-go-middlewares/request_id"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const basePath = "/api/cloud-connector/v1/"

// used to pass account, org_id down to request editor (to set headers)
type key int

const (
	accountKey key = iota
	orgIDKey   key = iota
)

type CloudConnectorClient interface {
	SendCloudConnectorRequest(
		ctx context.Context,
		account string,
		recipient uuid.UUID,
		url *string,
		directive string,
		metadata map[string]string,
	) (*string, bool, error)

	GetConnectionStatus(
		ctx context.Context,
		account string,
		orgID string,
		recipient string,
	) (ConnectionStatus, error)
}

type cloudConnectorClientImpl struct {
	client ClientWithResponsesInterface
}

func NewConnectorClientWithHttpRequestDoer(cfg *viper.Viper, doer HttpRequestDoer) CloudConnectorClient {
	client := &ClientWithResponses{
		ClientInterface: &Client{
			Server: fmt.Sprintf("%s://%s:%d%s", cfg.GetString("cloud.connector.scheme"), cfg.GetString("cloud.connector.host"), cfg.GetInt("cloud.connector.port"), basePath),
			Client: utils.NewMeasuredHttpRequestDoer(doer, "cloud-connector", "postMessage"),
			RequestEditor: func(ctx context.Context, req *http.Request) error {
				req.Header.Set(constants.HeaderRequestId, request_id.GetReqID(ctx))

				req.Header.Set(constants.HeaderCloudConnectorClientID, cfg.GetString("cloud.connector.client.id"))
				req.Header.Set(constants.HeaderCloudConnectorPSK, cfg.GetString("cloud.connector.psk"))
				req.Header.Set(constants.HeaderCloudConnectorAccount, ctx.Value(accountKey).(string))

				if orgID := ctx.Value(orgIDKey); orgID != nil {
					req.Header.Set(constants.HeaderCloudConnectorOrgID, ctx.Value(orgIDKey).(string))
				}

				// hotfix for the generated code escaping ampersands in string
				req.Body = utils.ReplaceAmpersand(req.Body)

				return nil
			},
		},
	}

	return &cloudConnectorClientImpl{
		client: client,
	}
}

func NewConnectorClient(cfg *viper.Viper) CloudConnectorClient {
	httpClient := http.Client{
		Timeout: time.Duration(cfg.GetInt64("cloud.connector.timeout") * int64(time.Second)),
	}

	return NewConnectorClientWithHttpRequestDoer(cfg, &httpClient)
}

func (this *cloudConnectorClientImpl) SendCloudConnectorRequest(
	ctx context.Context,
	account string,
	recipient uuid.UUID,
	url *string,
	directive string,
	metadata map[string]string,
) (id *string, notFound bool, err error) {
	ctx = context.WithValue(ctx, accountKey, account)
	recipientString := recipient.String()

	utils.GetLogFromContext(ctx).Debugw("Sending Cloud Connector message",
		"directive", directive,
		"metadata", metadata,
		"payload", url,
		"recipient", recipientString,
	)

	res, err := this.client.PostMessageWithResponse(ctx, PostMessageJSONRequestBody{
		Account:   &account,
		Directive: &directive,
		Metadata: &MessageRequest_Metadata{
			AdditionalProperties: metadata,
		},
		Payload:   url,
		Recipient: &recipientString,
	})

	if err != nil {
		return nil, false, err
	}

	if res.HTTPResponse.StatusCode == 404 {
		return nil, true, nil
	}

	if res.JSON201 == nil {
		return nil, false, unexpectedResponse(res.HTTPResponse)
	}

	return res.JSON201.Id, false, nil
}

func (this *cloudConnectorClientImpl) GetConnectionStatus(
	ctx context.Context,
	account string,
	orgID string,
	recipient string,
) (status ConnectionStatus, err error) {
	ctx = context.WithValue(ctx, accountKey, account)
	ctx = context.WithValue(ctx, orgIDKey, orgID)

	utils.GetLogFromContext(ctx).Debugw("Sending Cloud Connector status request",
		"account", account,
		"recipient", recipient,
	)

	res, err := this.client.V1ConnectionStatusMultiorgWithResponse(ctx, V1ConnectionStatusMultiorgJSONRequestBody{
		Account: &account,
		NodeId:  &recipient,
	})

	if err != nil {
		return "", err
	}

	if res.JSON200 == nil {
		return "", unexpectedResponse(res.HTTPResponse)
	}

	return *res.JSON200.Status, nil
}
