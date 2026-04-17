package constants

// Context key types for type-safe context value storage
type contextKey int

const (
	// Context keys (typed to avoid collisions)
	InternalRequestIdContextKey contextKey = iota
)

const (
	HeaderRequestId         = "x-rh-insights-request-id"
	HeaderInternalRequestId = "x-rh-playbook-dispatcher-internal-request-id"
	HeaderCorrelationId     = "x-rh-insights-playbook-dispatcher-correlation-id"
	HeaderIdentity          = "x-rh-identity"
	HeaderRequestType       = "service"

	HeaderCloudConnectorClientID = "x-rh-cloud-connector-client-id"
	HeaderCloudConnectorAccount  = "x-rh-cloud-connector-account"
	HeaderCloudConnectorPSK      = "x-rh-cloud-connector-psk"
	HeaderCloudConnectorOrgID    = "x-rh-cloud-connector-org-id"
)
