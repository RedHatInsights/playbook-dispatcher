package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	kafkaUtils "playbook-dispatcher/internal/common/kafka"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"strings"
	"time"

	"go.uber.org/zap"

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
	log      *zap.SugaredLogger
	producer *kafka.Producer
	schema   *jsonschema.Schema
	errors   chan error
}

func (this *handler) onMessage(msg *kafka.Message) {
	ctx := utils.SetLog(context.Background(), this.log)
	request := &messageModel.IngressValidationRequest{}
	err := json.Unmarshal(msg.Value, request)

	if err != nil {
		instrumentation.UnmarshallingError(ctx, err)
		return
	}

	// TODO: pass bounded logger
	ingressResponse, dispatcherResponse := this.handleRequest(ctx, request)

	this.produceMessage(ctx, ingressResponseTopic, ingressResponse)
	this.produceMessage(ctx, dispatcherResponseTopic, dispatcherResponse)
}

func (this *handler) handleRequest(ctx context.Context, request *messageModel.IngressValidationRequest) (*messageModel.IngressValidationResponse, *messageModel.PlaybookRunResponseMessageYaml) {
	ctx, _ = utils.WithRequestId(ctx, request.RequestID)

	if request.Size > cfg.GetInt64("artifact.max.size") {
		instrumentation.FileTooLarge(ctx, request)

		return messageModel.NewResponse(request, "failure"), nil
	}

	res, err := utils.DoGetWithRetry(client, request.URL, cfg.GetInt("storage.retries"))
	if err != nil {
		instrumentation.FetchArchiveError(ctx, err)
		return nil, nil
	}

	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	utils.GetLogFromContext(ctx).Debugw("Processing request", "account", request.Account, "reqId", request.RequestID)

	response := &messageModel.IngressValidationResponse{
		IngressValidationRequest: *request,
	}

	events, err := this.validateContent(data)
	if err != nil {
		response.Validation = validationFailure
		instrumentation.ValidationFailed(ctx, err)
		return response, nil
	}

	response.Validation = validationSuccess
	instrumentation.ValidationSuccess(ctx)

	dispatcherResponse := &messageModel.PlaybookRunResponseMessageYaml{
		Account:         request.Account,
		B64Identity:     request.B64Identity,
		RequestId:       request.RequestID,
		UploadTimestamp: request.Timestamp.Format(time.RFC3339),
		Events:          events,
	}

	return response, dispatcherResponse
}

func (this *handler) validateContent(data []byte) (events []messageModel.PlaybookRunResponseMessageYamlEventsElem, err error) {
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		errors, parserError := this.schema.ValidateBytes(context.TODO(), []byte(line))
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

func (this *handler) produceMessage(ctx context.Context, topic string, value interface{}) {
	if value != nil {
		if err := kafkaUtils.Produce(this.producer, topic, value); err != nil {
			instrumentation.ProducerError(ctx, err, topic) // TODO: request id
			this.errors <- err                             // TODO: is "shutdown-on-error" a good strategy?
		}
	}
}
