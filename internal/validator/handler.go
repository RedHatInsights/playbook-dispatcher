package validator

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/constants"
	commonInstrumentation "playbook-dispatcher/internal/common/instrumentation"
	kafkaUtils "playbook-dispatcher/internal/common/kafka"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"strings"
	"sync"
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
	timerFactory            = commonInstrumentation.OutboundHTTPDurationTimerFactory("storage")
	wg                      sync.WaitGroup
)

const (
	validationSuccess = "success"
	validationFailure = "failure"
)

type handler struct {
	producer     *kafka.Producer
	schema       *jsonschema.Schema
	errors       chan error
	requestsChan chan *messageModel.IngressValidationRequest
	validateChan chan messageInfo
}

func (this *handler) onMessage(ctx context.Context, msg *kafka.Message) {
	request := &messageModel.IngressValidationRequest{}
	err := json.Unmarshal(msg.Value, request)

	if err != nil {
		instrumentation.UnmarshallingError(ctx, err)
		return
	}

	ctx = utils.WithRequestId(ctx, request.RequestID)
	ctx = utils.WithAccount(ctx, request.Account)
	ctx = utils.SetLog(ctx, utils.GetLogFromContext(ctx).With("url", request.URL))
	utils.GetLogFromContext(ctx).Debugw("Processing request", "account", request.Account)

	this.requestsChan <- request
}

func (this *handler) initiateWorkers(
	ctx context.Context,
	workers int,
) {
	var workersWg sync.WaitGroup

	for i := 0; i < workers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()

			for {
				select {
				case msg, open := <-this.requestsChan:
					if !open {
						return
					}
					this.handleRequest(ctx, msg)
				}
			}
		}()
	}
	workersWg.Wait()
	close(this.validateChan)
}

func (this *handler) handleRequest(
	ctx context.Context,
	request *messageModel.IngressValidationRequest,
) {
	ingressResponse := &messageModel.IngressValidationResponse{
		IngressValidationRequest: *request,
	}

	if err := this.validateRequest(request); err != nil {
		this.validationFailed(ctx, err, ingressResponse)
	}

	res, err := utils.DoGetWithRetry(client, request.URL, cfg.GetInt("storage.retries"), timerFactory)
	if err != nil {
		instrumentation.FetchArchiveError(ctx, err)
		return
	}

	defer res.Body.Close()
	data, err := this.readFile(res.Body)
	if err != nil {
		this.validationFailed(ctx, err, ingressResponse)
		return
	}

	this.validateChan <- messageInfo{Request: request, Data: data, Response: ingressResponse}
}

func (this *handler) validationProcess(
	ctx context.Context,
	validateWg *sync.WaitGroup,
) {
	defer validateWg.Done()

	for {
		select {
		case chData, open := <-this.validateChan:
			if !open {
				return
			}
			this.validationSteps(ctx, chData)
		}

	}
}

func (this *handler) validationSteps(
	ctx context.Context,
	chData messageInfo,
) {
	request, ingressResponse, data := chData.Request, chData.Response, chData.Data

	events, err := this.validateContent(ctx, data)
	if err != nil {
		this.validationFailed(ctx, err, ingressResponse)
		utils.GetLogFromContext(ctx).Debugw("Invalid payload details", "data", string(data))
		return
	}

	correlationId, err := messageModel.GetCorrelationId(events)
	if err != nil {
		this.validationFailed(ctx, err, ingressResponse)
		return
	}

	ctx = utils.WithCorrelationId(ctx, correlationId.String())

	ingressResponse.Validation = validationSuccess
	instrumentation.ValidationSuccess(ctx)
	this.produceMessage(ctx, ingressResponseTopic, ingressResponse, request.Account)

	headers := kafkaUtils.Headers(constants.HeaderRequestId, request.RequestID, constants.HeaderCorrelationId, correlationId.String())
	dispatcherResponse := &messageModel.PlaybookRunResponseMessageYaml{
		Account:         request.Account,
		B64Identity:     request.B64Identity,
		RequestId:       request.RequestID,
		UploadTimestamp: request.Timestamp.Format(time.RFC3339),
		Events:          events,
	}

	this.produceMessage(ctx, dispatcherResponseTopic, dispatcherResponse, correlationId.String(), headers...)
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

func (this *handler) readFile(reader io.Reader) (result []byte, err error) {
	var isGzip bool
	reader = bufio.NewReaderSize(reader, 2)

	if isGzip, err = utils.IsGzip(reader); err != nil {
		return
	} else if isGzip {
		if gzipReader, err := gzip.NewReader(reader); err != nil {
			return nil, err
		} else {
			defer gzipReader.Close()
			reader = gzipReader
		}
	}

	return ioutil.ReadAll(reader)
}

type messageInfo struct {
	Request  *messageModel.IngressValidationRequest
	Response *messageModel.IngressValidationResponse
	Data     []byte
}
