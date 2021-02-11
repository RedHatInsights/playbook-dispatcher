package responseConsumer

import (
	"context"
	"io/ioutil"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/response-consumer/instrumentation"
	"sync"

	"github.com/spf13/viper"

	"github.com/ghodss/yaml"
	"github.com/qri-io/jsonschema"
)

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
) {
	instrumentation.Start()

	var schema jsonschema.Schema
	file, err := ioutil.ReadFile(cfg.GetString("schema.message.response"))
	utils.DieOnError(err)
	err = yaml.Unmarshal(file, &schema)
	utils.DieOnError(err)

	db, sql := db.Connect(ctx, cfg)
	ready.Register(sql.Ping)
	live.Register(sql.Ping)

	kafkaTimeout := cfg.GetInt("kafka.timeout")
	consumer, err := kafka.NewConsumer(ctx, cfg, cfg.GetString("topic.updates"))
	utils.DieOnError(err)

	ready.Register(func() error {
		return kafka.Ping(kafkaTimeout, consumer)
	})

	handler := &handler{
		db: db,
	}

	start := kafka.NewConsumerEventLoop(ctx, consumer, &schema, handler.onMessage, errors)

	go func() {
		defer wg.Done()
		defer utils.GetLogFromContext(ctx).Debug("Response consumer stopped")
		defer sql.Close()
		defer consumer.Close()
		wg.Add(1)
		start()
	}()
}
