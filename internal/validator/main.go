package validator

import (
	"context"
	"playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"sync"

	"github.com/spf13/viper"
)

const (
	payloadTypeHeader             = "service"
	playbookPayloadHeaderValue    = "playbook"
	playbookSatPayloadHeaderValue = "playbook-sat"
)

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
) {
	var schemaNames = []string{"schema.runner.event", "schema.rhcsat.event"}
	schemas := utils.LoadSchemas(cfg, schemaNames)

	storageConnectorConcurrency := cfg.GetInt("storage.max.concurrency")
	kafkaTimeout := cfg.GetInt("kafka.timeout")
	consumedTopic := cfg.GetString("topic.validation.request")
	consumer, err := kafka.NewConsumer(ctx, cfg, consumedTopic)
	utils.DieOnError(err)
	producer, err := kafka.NewProducer(cfg)
	utils.DieOnError(err)

	instrumentation.Start(cfg)

	handler := &handler{
		producer:     producer,
		schemas:      schemas,
		requestsChan: make(chan messageContext),
		validateChan: make(chan enrichedMessageContext),
	}

	storageConnector := newStorageConnector(cfg)
	var validateWg sync.WaitGroup

	ready.Register(func() error {
		return kafka.Ping(kafkaTimeout, consumer, producer)
	})

	predicate := kafka.FilterByHeaderPredicate(utils.GetLogFromContext(ctx), payloadTypeHeader, playbookPayloadHeaderValue, playbookSatPayloadHeaderValue)

	start := kafka.NewConsumerEventLoop(ctx, consumer, predicate, nil, handler.onMessage, errors)

	go func() {
		defer wg.Done()
		defer utils.GetLogFromContext(ctx).Debug("Validator stopped")
		defer producer.Close()
		defer utils.GetLogFromContext(ctx).Infof("Producer flushed with %d pending messages", producer.Flush(kafkaTimeout))
		defer validateWg.Wait()
		defer close(handler.requestsChan)
		defer consumer.Close()

		wg.Add(1)
		validateWg.Add(1)

		go storageConnector.initiateFetchWorkers(storageConnectorConcurrency, handler.requestsChan, handler.validateChan)
		go handler.initiateValidationWorker(&validateWg)

		start()
	}()
}
