package instrumentation

import (
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"

	dbModel "playbook-dispatcher/internal/common/model/db"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelParseUuid         = "parse_uuid"
	labelCorrelationId     = "correlation_id"
	labelMessageId         = "message_id"
	labelDb                = "db"
	labelPlaybookRunCreate = "playbook_run_create"
	labelPlaybookRunRead   = "playbook_run_read"
)

var (
	validationFailureTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_validation_failure_total",
		Help: "The total number of invalid requests",
	}, []string{"type", "subtype"})

	errorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_error_total",
		Help: "The total number of errors",
	}, []string{"type", "subtype"})

	connectorErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_cloud_connector_error_total",
		Help: "The total number of errors talking to cloud connector",
	})

	connectorSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_cloud_connector_sent_total",
		Help: "The total number of messages sent via cloud connector",
	})
)

func InvalidRecipientId(ctx echo.Context, value string, err error) {
	utils.GetLogFromEcho(ctx).Errorw("Error parsing recipient id", "error", err, "value", value)
	validationFailureTotal.WithLabelValues(labelParseUuid, labelCorrelationId).Inc()
}

func InvalidMessageId(ctx echo.Context, value string, err error) {
	utils.GetLogFromEcho(ctx).Errorw("Error parsing message id", "error", err, "value", value)
	validationFailureTotal.WithLabelValues(labelParseUuid, labelMessageId).Inc()
}

func CloudConnectorRequestError(ctx echo.Context, err error, recipient uuid.UUID) {
	utils.GetLogFromEcho(ctx).Errorw("Error sending message to cloud connector", "error", err, "recipient", recipient)
	connectorErrorTotal.Inc()
}

func CloudConnectorOK(ctx echo.Context, recipient uuid.UUID) {
	utils.GetLogFromEcho(ctx).Debugw("Received response from cloud connector", "recipient", recipient)
	connectorSentTotal.Inc()
}

func PlaybookRunCreateError(ctx echo.Context, err error, run *dbModel.Run) {
	utils.GetLogFromEcho(ctx).Errorw("Error creating run", "error", err, "run", *run)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate).Inc()
}

func PlaybookRunReadError(ctx echo.Context, err error) {
	utils.GetLogFromEcho(ctx).Errorw("Error reading playbook runs from database", "error", err)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead).Inc()
}

func Start() {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	validationFailureTotal.WithLabelValues(labelParseUuid, labelCorrelationId)
	validationFailureTotal.WithLabelValues(labelParseUuid, labelMessageId)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead)
}
