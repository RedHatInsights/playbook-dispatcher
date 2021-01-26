package validator

import (
	"context"
	"io/ioutil"
	"playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"

	"github.com/ghodss/yaml"
	"github.com/qri-io/jsonschema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Start(cfg *viper.Viper, log *zap.SugaredLogger, errors chan error, ready, live *utils.ProbeHandler) func(ctx context.Context) {
	var schema jsonschema.Schema
	file, err := ioutil.ReadFile(cfg.GetString("schema.runner.event"))
	utils.DieOnError(err)
	err = yaml.Unmarshal(file, &schema)
	utils.DieOnError(err)

	kafkaTimeout := cfg.GetInt("kafka.timeout")
	consumedTopic := cfg.GetString("topic.validation.request")
	consumer, err := kafka.NewConsumer(cfg, consumedTopic)
	utils.DieOnError(err)
	producer, err := kafka.NewProducer(cfg)
	utils.DieOnError(err)

	probes := instrumentation.Start(cfg, log)

	handler := &handler{
		log:      log,
		producer: producer,
		schema:   &schema,
		probes:   probes,
	}

	ready.Register(func() error {
		return kafka.Ping(kafkaTimeout, consumer, producer)
	})

	startLoop, stopLoop := kafka.NewConsumerEventLoop(consumer, nil, handler.onMessage, log)

	go startLoop()

	return func(ctx context.Context) {
		stopLoop()
		consumer.Close()
		producer.Flush(kafkaTimeout)
		producer.Close()
	}
}
