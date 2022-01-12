package instrumentation

import (
	"context"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"

	dbModel "playbook-dispatcher/internal/common/model/db"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelParseUuid             = "parse_uuid"
	labelCorrelationId         = "correlation_id"
	labelMessageId             = "message_id"
	labelDb                    = "db"
	labelPlaybookRunCreate     = "playbook_run_create"
	labelPlaybookRunHostCreate = "playbook_run_host_create"
	labelPlaybookRunRead       = "playbook_run_read"
	labelNoConnection          = "no_connection"
	labelErrorGeneric          = "error"
	labelAnsibleRequest        = "ansible"
	labelSatRequest            = "satellite"
)

var (
	validationFailureTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_validation_failure_total",
		Help: "The total number of invalid requests",
	}, []string{"type", "subtype", "request"})

	errorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_error_total",
		Help: "The total number of errors",
	}, []string{"type", "subtype", "request"})

	connectorErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_cloud_connector_error_total",
		Help: "The total number of errors talking to cloud connector",
	}, []string{"type", "request"})

	connectorSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_cloud_connector_sent_total",
		Help: "The total number of messages sent via cloud connector",
	})

	rbacErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_rbac_error_total",
		Help: "The total number of errors from RBAC",
	})

	rbacRejectedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_rbac_rejected_total",
		Help: "The total number of requests rejected due to RBAC",
	})

	runCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_run_created_total",
		Help: "The total number of created playbook runs",
	}, []string{"dispatching_service", "request"})
)

func InvalidRecipientId(ctx echo.Context, value string, err error, requestType string) {
	utils.GetLogFromEcho(ctx).Errorw("Error parsing recipient id", "error", err, "value", value)
	validationFailureTotal.WithLabelValues(labelParseUuid, labelCorrelationId, requestType).Inc()
}

func CloudConnectorRequestError(ctx context.Context, err error, recipient uuid.UUID, requestType string) {
	utils.GetLogFromContext(ctx).Errorw("Error sending message to cloud connector", "error", err, "recipient", recipient)
	connectorErrorTotal.WithLabelValues(labelErrorGeneric, requestType).Inc()
}

func CloudConnectorNoConnection(ctx context.Context, recipient uuid.UUID, requestType string) {
	utils.GetLogFromContext(ctx).Errorw("Cloud connector reporting no connection for recipient", "recipient", recipient)
	connectorErrorTotal.WithLabelValues(labelNoConnection, requestType).Inc()
}

func CloudConnectorOK(ctx context.Context, recipient uuid.UUID, messageId *string) {
	utils.GetLogFromContext(ctx).Debugw("Received response from cloud connector", "recipient", recipient, "message_id", *messageId)
	connectorSentTotal.Inc()
}

func PlaybookRunCreateError(ctx context.Context, err error, run *dbModel.Run, requestType string) {
	utils.GetLogFromContext(ctx).Errorw("Error creating run", "error", err, "run", *run)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate, requestType).Inc()
}

func PlaybookRunHostCreateError(ctx context.Context, err error, data []dbModel.RunHost, requestType string) {
	utils.GetLogFromContext(ctx).Errorw("Error creating run host", "error", err, "data", data)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunHostCreate, requestType).Inc()
}

func PlaybookRunReadError(ctx echo.Context, err error) {
	utils.GetLogFromEcho(ctx).Errorw("Error reading playbook runs from database", "error", err)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead).Inc()
}

func RbacError(ctx echo.Context, err error) {
	utils.GetLogFromEcho(ctx).Errorw("error getting permissions from RBAC", "error", err)
	rbacErrorTotal.Inc()
}

func RbacRejected(ctx echo.Context) {
	utils.GetLogFromEcho(ctx).Infow("access rejected due to RBAC")
	rbacRejectedTotal.Inc()
}

func RunCreated(ctx context.Context, recipient uuid.UUID, runId uuid.UUID, payload string, service string, requestType string) {
	utils.GetLogFromContext(ctx).Infow("Created new playbook run", "recipient", recipient.String(), "run_id", runId.String(), "payload", string(payload), "service", service)
	runCreatedTotal.WithLabelValues(service, requestType).Inc()
}

func Start() {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	validationFailureTotal.WithLabelValues(labelParseUuid, labelCorrelationId, labelAnsibleRequest)
	validationFailureTotal.WithLabelValues(labelParseUuid, labelMessageId, labelAnsibleRequest)
	validationFailureTotal.WithLabelValues(labelParseUuid, labelCorrelationId, labelSatRequest)
	validationFailureTotal.WithLabelValues(labelParseUuid, labelMessageId, labelSatRequest)

	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate, labelAnsibleRequest)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunHostCreate, labelAnsibleRequest)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead, labelAnsibleRequest)

	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate, labelSatRequest)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunHostCreate, labelSatRequest)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead, labelSatRequest)

	connectorErrorTotal.WithLabelValues(labelErrorGeneric, labelAnsibleRequest)
	connectorErrorTotal.WithLabelValues(labelErrorGeneric, labelSatRequest)
	connectorErrorTotal.WithLabelValues(labelNoConnection, labelAnsibleRequest)
	connectorErrorTotal.WithLabelValues(labelNoConnection, labelSatRequest)
}
