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
	errors       chan<- error
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
		OrgId:           request.OrgID,
		B64Identity:     request.B64Identity,
		RequestId:       request.RequestID,
		UploadTimestamp: request.Timestamp.Format(time.RFC3339),
		Events:          events.Playbook,
	}

	this.produceMessage(ctx, dispatcherResponseTopic, dispatcherResponse, correlationId.String(), headers...)
}

func (this *handler) validateRequest(request *messageModel.IngressValidationRequest) (err error) {
	if request.Size == 0 || request.Size > cfg.GetInt64("artifact.max.size") {
		return fmt.Errorf("Rejecting payload due to file size: %d", request.Size)
	}

	return
}

func (this *handler) validateContent(ctx context.Context, requestType string, data []byte) (events *messageModel.ValidatedMessages, err error) {
	events = &messageModel.ValidatedMessages{}
	events.PlaybookType = requestType

	log := utils.GetLogFromContext(ctx)

	maxMessageSize := 1 * 1024 * 1024
	maxStdoutSize := 1024

	// FIXME:  make this configurable
	truncateData := len(data) >= maxMessageSize
	if truncateData {
		log.Debug("Payload too big.  Truncating payload.")
	}

	var truncated string = "Truncated..."

	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		if requestType == playbookSatPayloadHeaderValue {
			validatedEvent, err := validateSatRunResponseWithSchema(ctx, this.schemas[1], line)

			if err == nil {
				err = validateSatHostUUID(validatedEvent)
			}

			if err != nil {
				return nil, err
			}

			if truncateData {
				// There could be one big console string
				if validatedEvent.Console != nil && len(*validatedEvent.Console) > maxStdoutSize {
					*validatedEvent.Console = (*validatedEvent.Console)[0:maxStdoutSize] + "..."
				}

				// There could also be too many console strings
				if i > 500 {
					if validatedEvent.Console != nil || *validatedEvent.Console != "" {
						validatedEvent.Console = &truncated
						truncated = ""
					}
				}
			}

			events.PlaybookSat = append(events.PlaybookSat, *validatedEvent)

		} else {
			validatedEvent, err := validateRunResponseWithSchema(ctx, this.schemas[0], line)
			if err != nil {
				return nil, err
			}

			if truncateData {
				// There could be one big stdout
				if validatedEvent.Stdout != nil && len(*validatedEvent.Stdout) > maxStdoutSize {
					*validatedEvent.Stdout = (*validatedEvent.Stdout)[0:maxStdoutSize] + "..."
				}

				// There could also be too many stdouts
				if i > 500 && i < len(lines)-2 {
					validatedEvent.Stdout = &truncated
					truncated = ""
				}
			}

			events.Playbook = append(events.Playbook, *validatedEvent)
		}
	}

	if len(events.PlaybookSat) == 0 && len(events.Playbook) == 0 {
		return nil, fmt.Errorf("No events found")
	}

	return events, nil
}

func validateSatHostUUID(event *messageModel.PlaybookSatRunResponseMessageYamlEventsElem) (err error) {
	if event.Host != nil {
		_, err = uuid.Parse(*event.Host)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateRunResponseWithSchema(ctx context.Context, schema *jsonschema.Schema, line string) (validatedEvent *messageModel.PlaybookRunResponseMessageYamlEventsElem, err error) {

	errors, parserError := schema.ValidateBytes(ctx, []byte(line))
	if parserError != nil {
		return nil, parserError
	} else if len(errors) > 0 {
		return nil, errors[0]
	}

	event := &messageModel.PlaybookRunResponseMessageYamlEventsElem{}
	err = json.Unmarshal([]byte(line), &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func validateSatRunResponseWithSchema(ctx context.Context, schema *jsonschema.Schema, line string) (validatedEvent *messageModel.PlaybookSatRunResponseMessageYamlEventsElem, err error) {

	errors, parserError := schema.ValidateBytes(ctx, []byte(line))
	if parserError != nil {
		return nil, parserError
	} else if len(errors) > 0 {
		return nil, errors[0]
	}

	event := &messageModel.PlaybookSatRunResponseMessageYamlEventsElem{}
	err = json.Unmarshal([]byte(line), &event)
	if err != nil {
		return nil, err
	}

	return event, nil
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

			if ignoreKafkaProduceError(err) {
				return
			}

			this.errors <- err // TODO: is "shutdown-on-error" a good strategy?
		}
	}
}

func ignoreKafkaProduceError(err error) bool {

	kafkaErr := err.(kafka.Error)

	return kafkaErr.Code() == kafka.ErrMsgSizeTooLarge
}
