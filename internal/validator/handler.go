package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/constants"
	kafkaUtils "playbook-dispatcher/internal/common/kafka"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/qri-io/jsonschema"
)

var (
	cfg    = config.Get()
	client = &http.Client{
		Timeout: time.Duration(cfg.GetInt64("storage.timeout") * int64(time.Second)),
	}
	ingressResponseTopic    = cfg.GetString("topic.validation.response")
	dispatcherResponseTopic = cfg.GetString("topic.updates")
)

const (
	validationSuccess = "success"
	validationFailure = "failure"
)

type handler struct {
	producer *kafka.Producer
	schema   *jsonschema.Schema
	errors   chan error
}

func (this *handler) onMessage(ctx context.Context, msg *kafka.Message) {
	request := &messageModel.IngressValidationRequest{}
	err := json.Unmarshal(msg.Value, request)

	if err != nil {
		instrumentation.UnmarshallingError(ctx, err)
		return
	}

	this.handleRequest(ctx, request)
}

func (this *handler) handleRequest(
	ctx context.Context,
	request *messageModel.IngressValidationRequest,
) {
	ctx = utils.WithRequestId(ctx, request.RequestID)
	utils.GetLogFromContext(ctx).Debugw("Processing request", "account", request.Account)

	ingressResponse := &messageModel.IngressValidationResponse{
		IngressValidationRequest: *request,
	}

	if err := this.validateRequest(request); err != nil {
		this.validationFailed(ctx, err, ingressResponse)
	}

	res, err := utils.DoGetWithRetry(client, request.URL, cfg.GetInt("storage.retries"))
	if err != nil {
		instrumentation.FetchArchiveError(ctx, err)
		return
	}

	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	events, err := this.validateContent(ctx, data)
	if err != nil {
		this.validationFailed(ctx, err, ingressResponse)
		return
	}

	correlationId, err := messageModel.GetCorrelationId(events)
	if err != nil {
		this.validationFailed(ctx, err, ingressResponse)
		return
	}

	ingressResponse.Validation = validationSuccess
	instrumentation.ValidationSuccess(ctx)
	this.produceMessage(ctx, ingressResponseTopic, ingressResponse, nil)

	headers := kafkaUtils.Headers(constants.HeaderRequestId, request.RequestID, constants.HeaderCorrelationId, correlationId.String())
	dispatcherResponse := &messageModel.PlaybookRunResponseMessageYaml{
		Account:         request.Account,
		B64Identity:     request.B64Identity,
		RequestId:       request.RequestID,
		UploadTimestamp: request.Timestamp.Format(time.RFC3339),
		Events:          events,
	}

	this.produceMessage(ctx, dispatcherResponseTopic, dispatcherResponse, nil, headers...)
}

func (this *handler) validateRequest(request *messageModel.IngressValidationRequest) (err error) {
	if request.Size > cfg.GetInt64("artifact.max.size") {
		return fmt.Errorf("Rejecting payload due to file size: %d", request.Size)
	}

	return
}

func (this *handler) validateContent(ctx context.Context, data []byte) (events []messageModel.PlaybookRunResponseMessageYamlEventsElem, err error) {
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		errors, parserError := this.schema.ValidateBytes(ctx, []byte(line))
		if parserError != nil {
			return nil, parserError
		} else if len(errors) > 0 {
			return nil, errors[0]
		}

		event := messageModel.PlaybookRunResponseMessageYamlEventsElem{}
		err = json.Unmarshal([]byte(line), &event)

		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("No events found")
	}

	return events, nil
}

func (this *handler) validationFailed(ctx context.Context, err error, response *messageModel.IngressValidationResponse) {
	response.Validation = validationFailure
	instrumentation.ValidationFailed(ctx, err)
	this.produceMessage(ctx, ingressResponseTopic, response, nil)
}

func (this *handler) produceMessage(ctx context.Context, topic string, value interface{}, key *string, headers ...kafka.Header) {
	if value != nil {
		if err := kafkaUtils.Produce(this.producer, topic, value, key, headers...); err != nil {
			instrumentation.ProducerError(ctx, err, topic)
			this.errors <- err // TODO: is "shutdown-on-error" a good strategy?
		}
	}
}
