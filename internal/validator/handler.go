package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/constants"
	kafkaUtils "playbook-dispatcher/internal/common/kafka"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/qri-io/jsonschema"
)

var (
	cfg                     = config.Get()
	ingressResponseTopic    = cfg.GetString("topic.validation.response")
	dispatcherResponseTopic = cfg.GetString("topic.updates")
)

const (
	validationSuccess = "success"
	validationFailure = "failure"
)

type handler struct {
	producer     *kafka.Producer
	schemas      []*jsonschema.Schema
	errors       chan error
	requestsChan chan messageContext
	validateChan chan enrichedMessageContext
}

type messageContext struct {
	requestType string
	request     messageModel.IngressValidationRequest
	ctx         context.Context
}

type enrichedMessageContext struct {
	data []byte
	messageContext
}

func (this *handler) onMessage(ctx context.Context, msg *kafka.Message) {
	request := messageModel.IngressValidationRequest{}
	requestType, _ := kafkaUtils.GetHeader(msg, payloadTypeHeader)

	ctx = utils.WithRequestType(ctx, requestType)

	err := json.Unmarshal(msg.Value, &request)

	if err != nil {
		instrumentation.UnmarshallingError(ctx, err, requestType)
		return
	}

	ctx = utils.WithRequestId(ctx, request.RequestID)
	ctx = utils.WithAccount(ctx, request.Account)
	ctx = utils.WithOrgId(ctx, request.OrgID)

	ctx = utils.SetLog(ctx, utils.GetLogFromContext(ctx).With("url", request.URL))
	utils.GetLogFromContext(ctx).Debugw("Processing request",
		"account", request.Account,
		"topic", *msg.TopicPartition.Topic,
		"partition", msg.TopicPartition.Partition,
		"offset", msg.TopicPartition.Offset.String(),
		"size", request.Size,
	)

	if err := this.validateRequest(&request); err != nil {
		this.validationFailed(ctx, err, requestType, &request)
		return
	}

	this.requestsChan <- messageContext{requestType: requestType, request: request, ctx: ctx}
}

func (this *handler) initiateValidationWorker(
	validateWg *sync.WaitGroup,
) {
	defer validateWg.Done()

	for {
		msg, open := <-this.validateChan

		if !open {
			return
		}
		this.validationSteps(msg)
	}
}

func (this *handler) validationSteps(
	msg enrichedMessageContext,
) {
	request, requestType, ctx, data := &msg.request, msg.requestType, msg.ctx, msg.data

	events, err := this.validateContent(ctx, requestType, data)
	if err != nil {
		this.validationFailed(ctx, err, requestType, request)
		utils.GetLogFromContext(ctx).Debugw("Invalid payload details", "data", string(data))
		return
	}

	correlationId, err := messageModel.GetCorrelationId(*events, playbookSatPayloadHeaderValue)
	if err != nil {
		this.validationFailed(ctx, err, requestType, request)
		return
	}

	ctx = utils.WithCorrelationId(ctx, correlationId.String())

	ingressResponse := &messageModel.IngressValidationResponse{
		IngressValidationRequest: *request,
		Validation:               validationSuccess,
	}

	instrumentation.ValidationSuccess(ctx, requestType)
	this.produceMessage(ctx, ingressResponseTopic, ingressResponse, request.Account)

	headers := kafkaUtils.Headers(constants.HeaderRequestId, request.RequestID, constants.HeaderCorrelationId, correlationId.String(), payloadTypeHeader, requestType)

	if requestType == playbookSatPayloadHeaderValue {
		dispatcherResponse := &messageModel.PlaybookSatRunResponseMessageYaml{
			Account:         request.Account,
			OrgId:           request.OrgID,
			B64Identity:     request.B64Identity,
			RequestId:       request.RequestID,
			UploadTimestamp: request.Timestamp.Format(time.RFC3339),
			Events:          events.PlaybookSat,
		}
		this.produceMessage(ctx, dispatcherResponseTopic, dispatcherResponse, correlationId.String(), headers...)
		return
	}

	dispatcherResponse := &messageModel.PlaybookRunResponseMessageYaml{
		Account:         request.Account,
		OrgId:           request.OrgID,
		B64Identity:     request.B64Identity,
		RequestId:       request.RequestID,
		UploadTimestamp: request.Timestamp.Format(time.RFC3339),
		Events:          events.Playbook,
	}

	this.produceMessage(ctx, dispatcherResponseTopic, dispatcherResponse, correlationId.String(), headers...)
}

func (this *handler) validateRequest(request *messageModel.IngressValidationRequest) (err error) {
	if request.Size > cfg.GetInt64("artifact.max.size") {
		return fmt.Errorf("Rejecting payload due to file size: %d", request.Size)
	}

	return
}

func (this *handler) validateContent(ctx context.Context, requestType string, data []byte) (events *messageModel.ValidatedMessages, err error) {
	events = &messageModel.ValidatedMessages{}
	events.PlaybookType = requestType

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		if requestType == playbookSatPayloadHeaderValue {
			err = validateWithSchema(ctx, this.schemas[1], true, line, events)
			if err == nil {
				err = validateSatHostUUID(line)
			}
		} else {
			err = validateWithSchema(ctx, this.schemas[0], false, line, events)
		}

		if err != nil {
			return nil, err
		}

	}

	if len(events.PlaybookSat) == 0 && len(events.Playbook) == 0 {
		return nil, fmt.Errorf("No events found")
	}

	return events, nil
}

func validateSatHostUUID(line string) (err error) {
	event := &messageModel.PlaybookSatRunResponseMessageYamlEventsElem{}
	err = json.Unmarshal([]byte(line), &event)

	if err != nil {
		return err
	}

	if event.Host != nil {
		_, err = uuid.Parse(*event.Host)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateWithSchema(ctx context.Context, schema *jsonschema.Schema, rhcsatRequest bool, line string, events *messageModel.ValidatedMessages) (err error) {
	errors, parserError := schema.ValidateBytes(ctx, []byte(line))
	if parserError != nil {
		return parserError
	} else if len(errors) > 0 {
		return errors[0]
	}

	if rhcsatRequest {
		event := &messageModel.PlaybookSatRunResponseMessageYamlEventsElem{}
		err = json.Unmarshal([]byte(line), &event)
		if err != nil {
			return err
		}

		events.PlaybookSat = append(events.PlaybookSat, *event)
		return
	}

	event := &messageModel.PlaybookRunResponseMessageYamlEventsElem{}
	err = json.Unmarshal([]byte(line), &event)
	if err != nil {
		return err
	}

	events.Playbook = append(events.Playbook, *event)
	return
}

func (this *handler) validationFailed(ctx context.Context, err error, requestType string, request *messageModel.IngressValidationRequest) {
	response := &messageModel.IngressValidationResponse{
		IngressValidationRequest: *request,
		Validation:               validationFailure,
	}

	instrumentation.ValidationFailed(ctx, err, requestType)
	this.produceMessage(ctx, ingressResponseTopic, response, response.Account)
}

func (this *handler) produceMessage(ctx context.Context, topic string, value interface{}, key string, headers ...kafka.Header) {
	if value != nil {
		if err := kafkaUtils.Produce(this.producer, topic, value, key, headers...); err != nil {
			instrumentation.ProducerError(ctx, err, topic)
			this.errors <- err // TODO: is "shutdown-on-error" a good strategy?
		}
	}
}
