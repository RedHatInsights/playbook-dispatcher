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
	playbook        = "playbook"
	playbookSat     = "playbook-sat"

	// Failure reasons are a fixed set so validator_failure_total can be
	// compared over time. The raw error stays in the log. ReasonUnknown is
	// the unclassified fallback; handler.failureReason says what to do when it climbs.
	ReasonSizeEmpty            = "size_empty"
	ReasonSizeExceeded         = "size_exceeded"
	ReasonSchema               = "schema"
	ReasonHostUUID             = "host_uuid"
	ReasonNoEvents             = "no_events"
	ReasonCorrelationIDMissing = "correlation_id_missing"
	ReasonCorrelationIDInvalid = "correlation_id_invalid"
	ReasonUnknown              = "unknown"
)

var failureReasons = []string{
	ReasonSizeEmpty,
	ReasonSizeExceeded,
	ReasonSchema,
	ReasonHostUUID,
	ReasonNoEvents,
	ReasonCorrelationIDMissing,
	ReasonCorrelationIDInvalid,
	ReasonUnknown,
}

var (
	validationSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validator_success_total",
		Help: "The total number of successfully validated payloads",
	}, []string{"request_type"})

	validationFailureTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validator_failure_total",
		Help: "The total number of payloads that did not pass validation",
	}, []string{"request_type", "reason"})

	errorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validator_error_total",
		Help: "The total number of errors during payloads processing",
	}, []string{"phase", "request_type"})

	producerError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validator_kafka_producer_error_total",
		Help: "The total number of kafka producer errors",
	}, []string{"topic"})
)

func Start(cfg *viper.Viper) {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	for _, requestType := range []string{playbook, playbookSat} {
		validationSuccessTotal.WithLabelValues(requestType)
		for _, reason := range failureReasons {
			validationFailureTotal.WithLabelValues(requestType, reason)
		}
		errorTotal.WithLabelValues(errorUnmarshall, requestType)
		errorTotal.WithLabelValues(errorS3, requestType)
	}
	producerError.WithLabelValues(cfg.GetString("topic.updates"))
	producerError.WithLabelValues(cfg.GetString("topic.validation.response"))
}

func ValidationSuccess(ctx context.Context, requestType string) {
	validationSuccessTotal.WithLabelValues(requestType).Inc()
	utils.GetLogFromContext(ctx).Infow("Payload valid")
}

func ValidationFailed(ctx context.Context, cause error, requestType string, reason string) {
	validationFailureTotal.WithLabelValues(requestType, reason).Inc()
	utils.GetLogFromContext(ctx).Warnw("Rejecting payload due to validation failure", "reason", reason, "cause", cause)
}

func UnmarshallingError(ctx context.Context, err error, requestType string) {
	errorTotal.WithLabelValues(errorUnmarshall, requestType).Inc()
	utils.GetLogFromContext(ctx).Errorw("Message unmarshalling failed", "error", err) // TODO some correlation info
}

func FetchArchiveError(ctx context.Context, err error, requestType string) {
	errorTotal.WithLabelValues(errorS3, requestType).Inc()
	utils.GetLogFromContext(ctx).Errorw("Failed to fetch uploaded archive", "error", err)
}

func ProducerError(ctx context.Context, err error, topic string) {
	producerError.WithLabelValues(topic).Inc()
	utils.GetLogFromContext(ctx).Errorw("Kafka producer error", "error", err, "topic", topic)
}
