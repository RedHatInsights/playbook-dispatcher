package instrumentation

import (
	"context"
	"playbook-dispatcher/internal/common/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
)

const (
	errorUnmarshall = "unmarshall"
	errorS3         = "s3fetch"
)

var (
	validationSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "validator_success_total",
		Help: "The total number of successfully validated payloads",
	})

	validationFailureTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "validator_failure_total",
		Help: "The total number of payloads that did not pass validation",
	})

	errorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validator_error_total",
		Help: "The total number of errors during payloads processing",
	}, []string{"phase"})

	producerError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validator_kafka_producer_error_total",
		Help: "The total number of kafka producer errors",
	}, []string{"topic"})
)

func Start(cfg *viper.Viper) {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	errorTotal.WithLabelValues(errorUnmarshall)
	errorTotal.WithLabelValues(errorS3)
	producerError.WithLabelValues(cfg.GetString("topic.updates"))
	producerError.WithLabelValues(cfg.GetString("topic.validation.response"))
}

func ValidationSuccess(ctx context.Context) {
	validationSuccessTotal.Inc()
	utils.GetLogFromContext(ctx).Debugw("Payload valid")
}

func ValidationFailed(ctx context.Context, cause error) {
	validationFailureTotal.Inc()
	utils.GetLogFromContext(ctx).Infow("Rejecting payload due to validation failure", "cause", cause)
}

func UnmarshallingError(ctx context.Context, err error) {
	errorTotal.WithLabelValues(errorUnmarshall).Inc()
	utils.GetLogFromContext(ctx).Errorw("Message unmarshalling failed", "error", err) // TODO some correlation info
}

func FetchArchiveError(ctx context.Context, err error) {
	errorTotal.WithLabelValues(errorS3).Inc()
	utils.GetLogFromContext(ctx).Errorw("Failed to fetch uploaded archive", "error", err)
}

func ProducerError(ctx context.Context, err error, topic string) {
	producerError.WithLabelValues(topic).Inc()
	utils.GetLogFromContext(ctx).Errorw("Kafka producer error", "error", err, "topic", topic)
}
