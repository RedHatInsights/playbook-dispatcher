package instrumentation

import (
	"context"
	"playbook-dispatcher/internal/common/utils"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	playbookRunUpdatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "response_consumer_playbook_run_updated_total",
		Help: "The total number of successfully updated playbook runs",
	})

	playbookRunUpdateMissTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "response_consumer_playbook_run_update_miss_total",
		Help: "The total number of run updates that did not match any known playbook run",
	})

	errorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "response_consumer_error_total",
		Help: "The total number of errors during payloads processing",
	}, []string{"type"})

	validationFailureTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "response_consumer_validation_failure_total",
		Help: "The total number of invalid payloads",
	}, []string{"type"})
)

const (
	labelDbUpdate       = "db_update"
	labelJsonUnmarshall = "json_unmarshall"
	labelHeaderMissing  = "header_missing"
)

func PlaybookRunUpdated(ctx context.Context, account string, status string, correlationId uuid.UUID) {
	utils.GetLogFromContext(ctx).Infow("Updated run", "account", account, "status", status, "correlation_id", correlationId)
	playbookRunUpdatedTotal.Inc()
}

func PlaybookRunUpdateMiss(ctx context.Context, account string, status string, correlationId uuid.UUID) {
	utils.GetLogFromContext(ctx).Warnw("No run to update", "account", account, "status", status, "correlation_id", correlationId)
	playbookRunUpdateMissTotal.Inc()
}

func PlaybookRunUpdateError(ctx context.Context, err error, account string, status string, correlationId uuid.UUID) {
	utils.GetLogFromContext(ctx).Errorw("Error updating run", "account", account, "status", status, "correlation_id", correlationId, "error", err)
	errorTotal.WithLabelValues(labelDbUpdate).Inc()
}

func UnmarshallIncomingMessageError(ctx context.Context, err error) {
	utils.GetLogFromContext(ctx).Errorw("Error unmarshalling incoming message", "error", err)
	validationFailureTotal.WithLabelValues(labelJsonUnmarshall).Inc()
}

func CannotReadHeaders(log *zap.SugaredLogger, err error) {
	log.Errorw("Error parsing correlation id", "error", err)
	errorTotal.WithLabelValues(labelHeaderMissing).Inc()
}

func Start() {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	errorTotal.WithLabelValues(labelDbUpdate)
	errorTotal.WithLabelValues(labelHeaderMissing)
	validationFailureTotal.WithLabelValues(labelJsonUnmarshall)
}
