package validator

import (
	"context"
	"io/ioutil"
	"playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"sync"

	k "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/ghodss/yaml"
	"github.com/qri-io/jsonschema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	payloadTypeHeader          = "service"
	playbookPayloadHeaderValue = "playbook"
)

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
) {
	var schema jsonschema.Schema
	file, err := ioutil.ReadFile(cfg.GetString("schema.runner.event"))
	utils.DieOnError(err)
	err = yaml.Unmarshal(file, &schema)
	utils.DieOnError(err)

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
		schema:       &schema,
		requestsChan: make(chan messageContext),
		validateChan: make(chan enrichedMessageContext),
	}

	storageConnector := newStorageConnector(cfg)
	var validateWg sync.WaitGroup

	ready.Register(func() error {
		return kafka.Ping(kafkaTimeout, consumer, producer)
	})

	predicate := newPlaybookServiceMessagePredicate(utils.GetLogFromContext(ctx))

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

func newPlaybookServiceMessagePredicate(log *zap.SugaredLogger) kafka.KafkaMessagePredicate {
	return func(msg *k.Message) bool {
		if val, err := kafka.GetHeader(msg, payloadTypeHeader); err != nil {
			log.Warnw("Error reading kafka message header", "err", err, "topic", *msg.TopicPartition.Topic, "partition", msg.TopicPartition.Partition, "offset", msg.TopicPartition.Offset.String())
			return false
		} else {
			return val == playbookPayloadHeaderValue
		}
	}
}
