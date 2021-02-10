package responseConsumer

import (
	"context"
	"io/ioutil"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/response-consumer/instrumentation"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/ghodss/yaml"
	"github.com/qri-io/jsonschema"
)

func Start(cfg *viper.Viper, log *zap.SugaredLogger, errors chan<- error, ready, live *utils.ProbeHandler) func(ctx context.Context) {
	instrumentation.Start()

	var schema jsonschema.Schema
	file, err := ioutil.ReadFile(cfg.GetString("schema.message.response"))
	utils.DieOnError(err)
	err = yaml.Unmarshal(file, &schema)
	utils.DieOnError(err)

	db, sql := db.Connect(cfg, log)
	ready.Register(sql.Ping)
	live.Register(sql.Ping)

	kafkaTimeout := cfg.GetInt("kafka.timeout")
	consumer, err := kafka.NewConsumer(cfg, cfg.GetString("topic.updates"))
	utils.DieOnError(err)

	ready.Register(func() error {
		return kafka.Ping(kafkaTimeout, consumer)
	})

	handler := &handler{
		db:  db,
		log: log,
	}

	startLoop, stopLoop := kafka.NewConsumerEventLoop(consumer, &schema, handler.onMessage, log)

	go startLoop()

	return func(ctx context.Context) {
		stopLoop()
		consumer.Close()
		sql.Close()
	}
}
