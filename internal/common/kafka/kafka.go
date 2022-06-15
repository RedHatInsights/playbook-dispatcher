package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"playbook-dispatcher/internal/common/utils"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/qri-io/jsonschema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var defaultTopic = "__consumer_offsets"

// https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md

func NewProducer(config *viper.Viper) (*kafka.Producer, error) {
	return kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":        config.GetString("kafka.bootstrap.servers"),
		"request.required.acks":    config.GetInt("kafka.request.required.acks"),
		"message.send.max.retries": config.GetInt("kafka.message.send.max.retries"),
		"retry.backoff.ms":         config.GetInt("kafka.retry.backoff.ms"),
	})
}

func NewConsumer(ctx context.Context, config *viper.Viper, topic string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        config.GetString("kafka.bootstrap.servers"),
		"group.id":                 config.GetString("kafka.group.id"),
		"auto.offset.reset":        config.GetString("kafka.auto.offset.reset"),
		"auto.commit.interval.ms":  config.GetInt("kafka.auto.commit.interval.ms"),
		"go.logs.channel.enable":   true,
		"allow.auto.create.topics": true,
	})

	if err != nil {
		return nil, err
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)

	if err != nil {
		return nil, err
	}

	go func() {
		log := utils.GetLogFromContext(ctx).Named("kafka")

		for {
			entry, ok := <-consumer.Logs()

			if !ok {
				return
			}

			log.Debug(entry)
		}
	}()

	return consumer, nil
}

func NewConsumerEventLoop(
	ctx context.Context,
	consumer *kafka.Consumer,
	messagePredicate KafkaMessagePredicate,
	validationPredicate KafkaMessagePredicate,
	handler func(context.Context, *kafka.Message),
	errors chan<- error,
) (start func()) {

	return func() {
		for {
			msg, err := consumer.ReadMessage(1 * time.Second) // TODO: configurable

			select {
			case <-ctx.Done():
				return
			default:
			}

			if err != nil {
				if err.(kafka.Error).Code() != kafka.ErrTimedOut {
					utils.GetLogFromContext(ctx).Errorw("Error reading message from kafka", "err", err)
					errors <- err
				}

				continue
			}

			if messagePredicate != nil && !messagePredicate(msg) {
				continue
			}

			if validationPredicate != nil && !validationPredicate(msg) {
				continue
			}

			handler(ctx, msg)
		}
	}
}

func Produce(producer *kafka.Producer, topic string, value interface{}, key string, headers ...kafka.Header) error {
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          marshalledValue,
		Key:            []byte(key),
	}

	if len(headers) > 0 {
		msg.Headers = headers
	}

	return producer.Produce(msg, nil)
}

type KafkaMessagePredicate func(msg *kafka.Message) bool

type pingable interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
}

func Ping(timeout int, instances ...pingable) error {
	for _, instance := range instances {
		if _, err := instance.GetMetadata(&defaultTopic, false, timeout); err != nil {
			return err
		}
	}

	return nil
}

func Headers(keysAndValues ...string) []kafka.Header {
	if len(keysAndValues)%2 != 0 {
		panic(fmt.Sprintf("Odd number of parameters: %s", keysAndValues))
	}

	result := make([]kafka.Header, len(keysAndValues))

	for i := 0; i < len(keysAndValues)/2; i++ {
		result[i] = kafka.Header{
			Key:   keysAndValues[i*2],
			Value: []byte(keysAndValues[(i*2)+1]),
		}
	}

	return result
}

func GetHeader(msg *kafka.Message, key string) (string, error) {
	for _, value := range msg.Headers {
		if value.Key == key {
			return string(value.Value), nil
		}
	}

	return "", fmt.Errorf("Header not found: %s", key)
}

func FilterByHeaderPredicate(log *zap.SugaredLogger, header string, filterVals ...string) KafkaMessagePredicate {
	return func(msg *kafka.Message) bool {
		if val, err := GetHeader(msg, header); err != nil {
			log.Warnw("Error reading kafka message header", "err", err, "topic", *msg.TopicPartition.Topic, "partition", msg.TopicPartition.Partition, "offset", msg.TopicPartition.Offset.String())
			return false
		} else {
			for _, filterVal := range filterVals {
				if val == filterVal {
					return true
				}
			}
			return false
		}
	}
}

func SchemaValidationPredicate(ctx context.Context, header string, schemaMapper map[string]*jsonschema.Schema) KafkaMessagePredicate {
	return func(msg *kafka.Message) bool {
		val, _ := GetHeader(msg, header)

		schema := schemaMapper[val]
		errors, parserError := schema.ValidateBytes(ctx, msg.Value)
		if len(errors) > 0 {
			utils.GetLogFromContext(ctx).Warnw("Incoming message does not match schema", "err", errors[0])
			return false
		} else if parserError != nil {
			utils.GetLogFromContext(ctx).Warnw("Incoming message cannot be parsed", "err", parserError)
			return false
		}
		return true
	}
}
