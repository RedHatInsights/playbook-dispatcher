package validator

import (
	"context"
	"io/ioutil"
	"playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/labstack/echo/v4"
	"github.com/qri-io/jsonschema"
	"github.com/spf13/viper"
)

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
	server *echo.Echo,
) {
	var schema jsonschema.Schema
	file, err := ioutil.ReadFile(cfg.GetString("schema.runner.event"))
	utils.DieOnError(err)
	err = yaml.Unmarshal(file, &schema)
	utils.DieOnError(err)

	kafkaTimeout := cfg.GetInt("kafka.timeout")
	consumedTopic := cfg.GetString("topic.validation.request")
	consumer, err := kafka.NewConsumer(ctx, cfg, consumedTopic)
	utils.DieOnError(err)
	producer, err := kafka.NewProducer(cfg)
	utils.DieOnError(err)

	instrumentation.Start(cfg)

	handler := &handler{
		producer: producer,
		schema:   &schema,
	}

	ready.Register(func() error {
		return kafka.Ping(kafkaTimeout, consumer, producer)
	})

	start := kafka.NewConsumerEventLoop(ctx, consumer, nil, handler.onMessage, errors)

	go func() {
		defer wg.Done()
		defer utils.GetLogFromContext(ctx).Debug("Validator stopped")
		defer producer.Close()
		defer utils.GetLogFromContext(ctx).Infof("Producer flushed with %d pending messages", producer.Flush(kafkaTimeout))
		defer consumer.Close()
		wg.Add(1)
		start()
	}()
}
