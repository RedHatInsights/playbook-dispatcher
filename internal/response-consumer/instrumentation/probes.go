package instrumentation

import (
	"context"
	"playbook-dispatcher/internal/common/utils"

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

	playbookSequenceOutOfOrder = promauto.NewCounter(prometheus.CounterOpts{
		Name: "response_consumer_playbook_run_sequence_out_of_order_total",
		Help: "The total number of run updates that are consumed out of order",
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

func PlaybookRunUpdated(ctx context.Context, status string, runId uuid.UUID) {
	utils.GetLogFromContext(ctx).Infow("Updated run", "runStatus", status, "run_id", runId.String())
	playbookRunUpdatedTotal.Inc()
}

func PlaybookRunUpdateMiss(ctx context.Context, status string) {
	utils.GetLogFromContext(ctx).Warnw("No run to update", "runStatus", status)
	playbookRunUpdateMissTotal.Inc()
}

func PlaybookRunUpdateError(ctx context.Context, err error, status string, runId uuid.UUID) {
	utils.GetLogFromContext(ctx).Errorw("Error updating run", "runStatus", status, "error", err, "run_id", runId.String())
	errorTotal.WithLabelValues(labelDbUpdate).Inc()
}

func UnmarshallIncomingMessageError(ctx context.Context, err error) {
	utils.GetLogFromContext(ctx).Errorw("Error unmarshalling incoming message", "error", err)
	validationFailureTotal.WithLabelValues(labelJsonUnmarshall).Inc()
}

func CannotReadHeaders(ctx context.Context, err error) {
	utils.GetLogFromContext(ctx).Errorw("Error parsing correlation id", "error", err)
	errorTotal.WithLabelValues(labelHeaderMissing).Inc()
}

func PlaybookRunUpdateSequenceOrder(ctx context.Context) {
	utils.GetLogFromContext(ctx).Errorw("Run update is out of order")
	playbookSequenceOutOfOrder.Inc()
}

func Start() {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	errorTotal.WithLabelValues(labelDbUpdate)
	errorTotal.WithLabelValues(labelHeaderMissing)
	validationFailureTotal.WithLabelValues(labelJsonUnmarshall)
}
