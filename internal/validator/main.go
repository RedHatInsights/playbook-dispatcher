package validator

import (
	"context"
	"io/ioutil"
	"log"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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

	/*
		kafkaTimeout := cfg.GetInt("kafka.timeout")
		consumedTopic := cfg.GetString("topic.validation.request")
		consumer, err := kafka.NewConsumer(ctx, cfg, consumedTopic)
		utils.DieOnError(err)
		producer, err := kafka.NewProducer(cfg)
		utils.DieOnError(err)
	*/

	instrumentation.Start(cfg)

	handler := &handler{
		schema: &schema,
	}

	/*
		ready.Register(func() error {
			return kafka.Ping(kafkaTimeout, consumer, producer)
		})
	*/

	//start := kafka.NewConsumerEventLoop(ctx, consumer, nil, handler.onMessage, errors)

	protocol, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	httpHandler, err := cloudevents.NewHTTPReceiveHandler(ctx, protocol, func(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {
		data := &messageModel.IngressValidationRequest{}
		if err := event.DataAs(data); err != nil {
			log.Printf("Error while extracting cloudevent Data: %s\n", err.Error())
			return nil, cloudevents.NewHTTPResult(400, "failed to convert data: %s", err)
		}

		utils.GetLogFromContext(ctx).Infow("Received cloudevent",
			"type", event.Type(),
			"specVersion", event.SpecVersion(),
			"source", event.Source(),
			"subject", event.Subject(),
			"id", event.ID(),
			"time", event.Time(),
		)

		handler.handleRequest(ctx, data)

		return nil, nil
	})
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}

	server.Any("/", echo.WrapHandler(httpHandler))

	go func() {
		defer wg.Done()
		defer utils.GetLogFromContext(ctx).Debug("Validator stopped")
		//defer producer.Close()
		//defer utils.GetLogFromContext(ctx).Infof("Producer flushed with %d pending messages", producer.Flush(kafkaTimeout))
		//defer consumer.Close()
		<-ctx.Done()
		wg.Add(1)

		//start()
	}()
}
