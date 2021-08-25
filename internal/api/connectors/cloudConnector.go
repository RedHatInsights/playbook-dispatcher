package connectors

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/constants"
	"time"

	"playbook-dispatcher/internal/common/utils"
	"strconv"

	"github.com/redhatinsights/platform-go-middlewares/request_id"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const basePath = "/api/cloud-connector/v1/"

var cloudConnectorDirective = "rhc-worker-playbook"

type accountKeyType int

// used to pass account down to request editor (to set headers)
const accountKey accountKeyType = iota

type CloudConnectorClient interface {
	SendCloudConnectorRequest(
		ctx context.Context,
		account string,
		recipient uuid.UUID,
		correlationId uuid.UUID,
		url string,
	) (*string, bool, error)
}

type cloudConnectorClientImpl struct {
	returnUrl        string
	responseInterval int
	client           ClientWithResponsesInterface
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
				return nil
			},
		},
	}

	return &cloudConnectorClientImpl{
		returnUrl:        cfg.GetString("return.url"),
		responseInterval: cfg.GetInt("response.interval"),
		client:           client,
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
	correlationId uuid.UUID,
	url string,
) (id *string, notFound bool, err error) {
	ctx = context.WithValue(ctx, accountKey, account)
	recipientString := recipient.String()
	metadata := map[string]string{
		"return_url":                    this.returnUrl,
		"response_interval":             strconv.Itoa(this.responseInterval),
		"crc_dispatcher_correlation_id": correlationId.String(),
	}

	utils.GetLogFromContext(ctx).Debugw("Sending Cloud Connector message",
		"directive", cloudConnectorDirective,
		"metadata", metadata,
		"payload", url,
		"recipient", recipientString,
	)

	res, err := this.client.PostMessageWithResponse(ctx, PostMessageJSONRequestBody{
		Account:   &account,
		Directive: &cloudConnectorDirective,
		Metadata: &MessageRequest_Metadata{
			AdditionalProperties: metadata,
		},
		Payload:   &url,
		Recipient: &recipientString,
	})

	if err != nil {
		return nil, false, err
	}

	if res.HTTPResponse.StatusCode == 404 {
		return nil, true, nil
	}

	if res.JSON201 == nil {
		return nil, false, fmt.Errorf(`unexpected status code "%d" or content type "%s"`, res.HTTPResponse.StatusCode, res.HTTPResponse.Header.Get("content-type"))
	}

	return res.JSON201.Id, false, nil
}
