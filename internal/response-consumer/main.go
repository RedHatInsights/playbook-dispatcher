package responseConsumer

import (
	"context"
	"io/ioutil"
	"log"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/response-consumer/instrumentation"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/labstack/echo/v4"
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
	server *echo.Echo,
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

	/*
		kafkaTimeout := cfg.GetInt("kafka.timeout")
		consumer, err := kafka.NewConsumer(ctx, cfg, cfg.GetString("topic.updates"))
		utils.DieOnError(err)

		ready.Register(func() error {
			return kafka.Ping(kafkaTimeout, consumer)
		})
	*/

	handler := &handler{
		db: db,
	}

	//start := kafka.NewConsumerEventLoop(ctx, consumer, &schema, handler.onMessage, errors)

	protocol, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	httpHandler, err := cloudevents.NewHTTPReceiveHandler(ctx, protocol, func(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {
		data := &message.PlaybookRunResponseMessageYaml{}
		if err := event.DataAs(data); err != nil {
			log.Printf("Error while extracting cloudevent Data: %s\n", err.Error())
			return nil, cloudevents.NewHTTPResult(400, "failed to convert data: %s", err)
		}

		handler.onRunResponse(ctx, data)

		return nil, nil
	})
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}

	server.Any("/", echo.WrapHandler(httpHandler))

	go func() {
		defer wg.Done()
		defer utils.GetLogFromContext(ctx).Debug("Response consumer stopped")
		defer sql.Close()
		<-ctx.Done()
		wg.Add(1)
	}()
}
