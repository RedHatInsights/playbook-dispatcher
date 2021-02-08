package connectors

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var cloudConnectorDirective = "playbook"

type CloudConnectorClient interface {
	SendCloudConnectorRequest(ctx context.Context, account string, recipient uuid.UUID) (*string, error)
}

type cloudConnectorClientImpl struct {
	log       *zap.SugaredLogger
	returnUrl string
	client    ClientWithResponsesInterface
}

func NewConnectorClientWithHttpRequestDoer(cfg *viper.Viper, log *zap.SugaredLogger, doer HttpRequestDoer) CloudConnectorClient {
	client := &ClientWithResponses{
		ClientInterface: &Client{
			Server: cfg.GetString("cloud.connector.host"),
			Client: doer,
			RequestEditor: func(ctx context.Context, req *http.Request) error {
				// TODO: forward request id
				return nil
			},
		},
	}

	return &cloudConnectorClientImpl{
		log:       log,
		returnUrl: cfg.GetString("return.url"),
		client:    client,
	}
}

func NewConnectorClient(cfg *viper.Viper, log *zap.SugaredLogger) CloudConnectorClient {
	httpClient := http.Client{
		Timeout: time.Duration(cfg.GetInt64("cloud.connector.timeout") * int64(time.Second)),
	}

	return NewConnectorClientWithHttpRequestDoer(cfg, log, &httpClient)
}

func (this *cloudConnectorClientImpl) SendCloudConnectorRequest(ctx context.Context, account string, recipient uuid.UUID) (*string, error) {
	recipientString := recipient.String()
	metadata := map[string]interface{}{
		"return_url": this.returnUrl,
	}

	// TODO: probe
	this.log.Debugw("Sending Cloud Connector message",
		"account", account,
		"directive", cloudConnectorDirective,
		"metadata", metadata,
		"recipient", recipientString,
	)

	res, err := this.client.PostMessageWithResponse(ctx, PostMessageJSONRequestBody{
		Account:   &account,
		Directive: &cloudConnectorDirective,
		Metadata:  &metadata,
		// TODO: payload? content url?
		Recipient: &recipientString,
	})

	if err != nil {
		return nil, err
	}

	if res.JSON200 == nil {
		return nil, fmt.Errorf(`unexpected status code "%d" or content type "%s"`, res.HTTPResponse.StatusCode, res.HTTPResponse.Header.Get("content-type"))
	}

	return res.JSON200.Id, nil
}
