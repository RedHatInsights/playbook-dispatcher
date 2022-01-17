package responseConsumer

import (
	"context"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/response-consumer/instrumentation"
	"sync"

	"github.com/qri-io/jsonschema"
	"github.com/spf13/viper"
)

const (
	requestTypeHeader        = "service"
	runnerMessageHeaderValue = "playbook"
	satMessageHeaderValue    = "playbook-sat"
)

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
) {
	instrumentation.Start()

	schemaMapper := make(map[string]*jsonschema.Schema)
	var schemaNames = []string{"schema.message.response", "schema.satmessage.response"}

	schemas := utils.LoadSchemas(cfg, schemaNames)
	schemaMapper[runnerMessageHeaderValue] = schemas[0]
	schemaMapper[satMessageHeaderValue] = schemas[1]

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

	headerPredicate := kafka.FilterByHeaderPredicate(utils.GetLogFromContext(ctx), requestTypeHeader, runnerMessageHeaderValue, satMessageHeaderValue)
	validationPredicate := kafka.SchemaValidationPredicate(ctx, requestTypeHeader, schemaMapper)

	start := kafka.NewConsumerEventLoop(ctx, consumer, headerPredicate, validationPredicate, handler.onMessage, errors)

	go func() {
		defer wg.Done()
		defer utils.GetLogFromContext(ctx).Debug("Response consumer stopped")
		defer sql.Close()
		defer consumer.Close()
		wg.Add(1)
		start()
	}()
}
