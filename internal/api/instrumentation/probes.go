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
	v1 = "v1"
	v2 = "v2"

	labelDb                    = "db"
	labelPlaybookRunCreate     = "playbook_run_create"
	labelPlaybookRunHostCreate = "playbook_run_host_create"
	labelPlaybookRunRead       = "playbook_run_read"
	labelNoConnection          = "no_connection"
	labelErrorGeneric          = "error"
	labelTenantAnemic          = "anemic-tenant"
	labelSatellite             = "satellite"
	LabelAnsibleRequest        = "ansible"
	LabelSatRequest            = "satellite"
)

var (
	validationFailureTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_validation_failure_total",
		Help: "The total number of invalid requests",
	}, []string{"type"})

	errorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_error_total",
		Help: "The total number of errors",
	}, []string{"type", "subtype", "request", "api_version"})

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
	}, []string{"dispatching_service", "request", "api_version"})

	runCanceledTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_run_canceled_total",
		Help: "The total number of canceled playbook runs",
	})

	runCanceledErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_run_canceled_error_total",
		Help: "The total number of errors from the run cancel endpoint",
	})
)

func TenantAnemic(ctx echo.Context, orgID string) {
	utils.GetLogFromEcho(ctx).Errorw("Rejecting request for anemic tenant", "org_id", orgID)
	validationFailureTotal.WithLabelValues(labelTenantAnemic).Inc()
}

func InvalidSatelliteRequest(ctx echo.Context, err error) {
	utils.GetLogFromEcho(ctx).Errorw("Invalid Satellite request", "error", err)
	validationFailureTotal.WithLabelValues(labelSatellite).Inc()
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
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate, requestType, utils.GetApiVersion(ctx)).Inc()
}

func PlaybookRunHostCreateError(ctx context.Context, err error, data []dbModel.RunHost, requestType string) {
	utils.GetLogFromContext(ctx).Errorw("Error creating run host", "error", err, "data", data)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunHostCreate, requestType, utils.GetApiVersion(ctx)).Inc()
}

func PlaybookRunCancelError(ctx context.Context, err error) {
	utils.GetLogFromContext(ctx).Errorw("Error canceling run", "error", err)
	runCanceledErrorTotal.Inc()
}

func PlaybookRunCancelRunTypeError(ctx context.Context, runId uuid.UUID) {
	utils.GetLogFromContext(ctx).Errorw("Attempting to cancel run not of type Satellite RHC")
	runCanceledErrorTotal.Inc()
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
	runCreatedTotal.WithLabelValues(service, requestType, utils.GetApiVersion(ctx)).Inc()
}

func RunCanceled(ctx context.Context, runId uuid.UUID) {
	utils.GetLogFromContext(ctx).Infow("Successfully initiated playbook run cancelation", "run_id", runId.String())
	runCanceledTotal.Inc()
}

func Start() {
	// initialize label values
	// https://www.robustperception.io/existential-issues-with-metrics
	validationFailureTotal.WithLabelValues(labelTenantAnemic)
	validationFailureTotal.WithLabelValues(labelSatellite)

	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate, LabelAnsibleRequest, v1)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunHostCreate, LabelAnsibleRequest, v1)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead, LabelAnsibleRequest, v1)

	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate, LabelAnsibleRequest, v2)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunHostCreate, LabelAnsibleRequest, v2)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead, LabelAnsibleRequest, v2)

	errorTotal.WithLabelValues(labelDb, labelPlaybookRunCreate, LabelSatRequest, v2)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunHostCreate, LabelSatRequest, v2)
	errorTotal.WithLabelValues(labelDb, labelPlaybookRunRead, LabelSatRequest, v2)

	connectorErrorTotal.WithLabelValues(labelErrorGeneric, LabelAnsibleRequest)
	connectorErrorTotal.WithLabelValues(labelErrorGeneric, LabelSatRequest)
	connectorErrorTotal.WithLabelValues(labelNoConnection, LabelAnsibleRequest)
	connectorErrorTotal.WithLabelValues(labelNoConnection, LabelSatRequest)
}
