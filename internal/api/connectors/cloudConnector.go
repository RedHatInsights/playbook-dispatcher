package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"playbook-dispatcher/internal/common/constants"
	"time"

	"playbook-dispatcher/internal/common/utils"

	"github.com/redhatinsights/platform-go-middlewares/request_id"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const basePath = "/api/cloud-connector/"

// used to pass account, org_id down to request editor (to set headers)
type key int

const (
	orgIDKey key = iota
)

type CloudConnectorClient interface {
	SendCloudConnectorRequest(
		ctx context.Context,
		orgID string,
		recipient uuid.UUID,
		url *string,
		directive string,
		metadata map[string]string,
	) (*string, bool, error)

	GetConnectionStatus(
		ctx context.Context,
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
				req.Header.Set(constants.HeaderCloudConnectorOrgID, ctx.Value(orgIDKey).(string))

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

func encodedBody(body PostV2ConnectionsClientIdMessageJSONRequestBody) (io.Reader, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(body); err != nil {
		return nil, err
	}
	return buf, nil
}

func (this *cloudConnectorClientImpl) SendCloudConnectorRequest(
	ctx context.Context,
	orgID string,
	recipient uuid.UUID,
	url *string,
	directive string,
	metadata map[string]string,
) (id *string, notFound bool, err error) {
	recipientString := recipient.String()

	ctx = context.WithValue(ctx, orgIDKey, orgID)

	utils.GetLogFromContext(ctx).Debugw("Sending Cloud Connector message",
		"directive", directive,
		"metadata", metadata,
		"payload", url,
		"recipient", recipientString,
	)

	body, err := encodedBody(PostV2ConnectionsClientIdMessageJSONRequestBody{
		Directive: &directive,
		Metadata: &MessageRequestV2_Metadata{
			AdditionalProperties: metadata,
		},
		Payload: url,
	})

	if err != nil {
		return nil, false, err
	}

	res, err := this.client.PostV2ConnectionsClientIdMessageWithBodyWithResponse(ctx, ClientID(recipientString), "application/json", body)

	if err != nil {
		return nil, false, err
	}

	if res.HTTPResponse.StatusCode == 404 {
		return nil, true, nil
	}

	if res.JSON201 == nil {
		return nil, false, utils.UnexpectedResponse(res.HTTPResponse)
	}

	return res.JSON201.Id, false, nil
}

func (this *cloudConnectorClientImpl) GetConnectionStatus(
	ctx context.Context,
	orgID string,
	recipient string,
) (status ConnectionStatus, err error) {
	ctx = context.WithValue(ctx, orgIDKey, orgID)

	utils.GetLogFromContext(ctx).Debugw("Sending Cloud Connector status request",
		"org_id", orgID,
		"recipient", recipient,
	)

	res, err := this.client.V2ConnectionStatusMultiorgWithResponse(ctx, ClientID(recipient))

	if err != nil {
		return "", err
	}

	return *res.JSON200.Status, nil
}
