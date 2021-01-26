package instrumentation

import (
	messageModel "playbook-dispatcher/internal/common/model/message"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	errorUnmarshall = "unmarshall"
	errorS3         = "s3fetch"
)

var (
	validationSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "validation_success_total",
		Help: "The total number of successfully validated payloads",
	})

	validationFailureTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "validation_failure_total",
		Help: "The total number of payloads that did not pass validation",
	})

	errorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validation_error_total",
		Help: "The total number of errors during payloads processing",
	}, []string{"phase"})

	producerError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_producer_error_total",
		Help: "The total number of kafka producer errors",
	}, []string{"topic"})
)

type Probes struct {
	log *zap.SugaredLogger
}

func Start(cfg *viper.Viper, log *zap.SugaredLogger) *Probes {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	errorTotal.WithLabelValues(errorUnmarshall)
	errorTotal.WithLabelValues(errorS3)
	producerError.WithLabelValues(cfg.GetString("topic.responses"))
	producerError.WithLabelValues(cfg.GetString("topic.validation.response"))

	return New(log)
}

func New(log *zap.SugaredLogger) *Probes {
	return &Probes{
		log: log,
	}
}

func (this *Probes) ValidationSuccess(request *messageModel.IngressValidationRequest) {
	validationSuccessTotal.Inc()
	this.log.Debugw("Payload valid", "reqId", request.RequestID)
}

func (this *Probes) ValidationFailed(request *messageModel.IngressValidationRequest, cause error) {
	validationFailureTotal.Inc()
	this.log.Infow("Rejecting payload due to validation failure", "cause", cause, "reqId", request.RequestID)
}

func (this *Probes) UnmarshallingError(err error) {
	errorTotal.WithLabelValues(errorUnmarshall).Inc()
	this.log.Errorw("Message unmarshalling failed", "error", err) // TODO some correlation info
}

func (this *Probes) FetchArchiveError(request *messageModel.IngressValidationRequest, err error) {
	errorTotal.WithLabelValues(errorS3).Inc()
	this.log.Errorw("Failed to fetch uploaded archive", "error", err, "reqId", request.RequestID)
}

func (this *Probes) FileTooLarge(request *messageModel.IngressValidationRequest) {
	validationFailureTotal.Inc()
	this.log.Infow("Rejecting payload due to file size", "size", request.Size, "reqId", request.RequestID)
}

func (this *Probes) ProducerError(err error, topic string) {
	producerError.WithLabelValues(topic).Inc()
	this.log.Errorw("Kafka producer error", "error", err, "topic", topic)
}
