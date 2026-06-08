package middleware

import (
	"net/http"
	"net/http/httptest"
	"playbook-dispatcher/internal/api/rbac"
	"playbook-dispatcher/internal/common/kessel"
	"playbook-dispatcher/internal/common/utils"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetRbacAllowedServices_Empty(t *testing.T) {
	permissions := []rbac.Access{}
	result := getRbacAllowedServices(permissions)
	assert.Empty(t, result)
}

func TestLogComparison_Match(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Set logger in request context
	req = req.WithContext(utils.SetLog(req.Context(), log))

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	rbacServices := []string{"config_manager", "remediations"}
	kesselServices := []string{"remediations", "config_manager"} // Different order

	// Should not panic
	logComparison(ctx, rbacServices, kesselServices, log)
}

func TestLogComparison_Mismatch(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Set logger in request context
	req = req.WithContext(utils.SetLog(req.Context(), log))

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	rbacServices := []string{"config_manager", "remediations"}
	kesselServices := []string{"config_manager", "tasks"}

	// Should not panic and should log mismatch
	logComparison(ctx, rbacServices, kesselServices, log)
}

func TestLogComparison_EmptyLists(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Set logger in request context
	req = req.WithContext(utils.SetLog(req.Context(), log))

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	rbacServices := []string{}
	kesselServices := []string{}

	// Should handle empty lists
	logComparison(ctx, rbacServices, kesselServices, log)
}

func TestLogComparison_OneEmpty(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Set logger in request context
	req = req.WithContext(utils.SetLog(req.Context(), log))

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	rbacServices := []string{"config_manager"}
	kesselServices := []string{}

	// Should handle mismatch when one is empty
	logComparison(ctx, rbacServices, kesselServices, log)
}

func TestExtractServicesToCheck_ReturnsImmutableCopy(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	// Request with no service filter - should return copy of all services
	req := httptest.NewRequest(http.MethodGet, "/api/playbook-dispatcher/v1/runs", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Get the original count
	originalCount := len(kessel.V2ApplicationPermissions)
	assert.Equal(t, 3, originalCount) // Ensure we start with 3 services

	// Get services from helper
	result := extractServicesToCheck(ctx, log)

	// Verify we got all services
	assert.Equal(t, 3, len(result))

	// Mutate the returned map
	result["malicious_service"] = "malicious_permission"

	// Verify the global map is unchanged
	assert.Equal(t, originalCount, len(kessel.V2ApplicationPermissions))
	_, exists := kessel.V2ApplicationPermissions["malicious_service"]
	assert.False(t, exists, "Global V2ApplicationPermissions should not be mutated")
}

func TestExtractServicesToCheck_WithFilter(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	// Request with service filter
	req := httptest.NewRequest(http.MethodGet, "/api/playbook-dispatcher/v1/runs?filter[service]=remediations", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	result := extractServicesToCheck(ctx, log)

	// Should return only the requested service
	assert.Equal(t, 1, len(result))
	assert.Contains(t, result, "remediations")
	assert.Equal(t, kessel.PermissionRemediationsRunView, result["remediations"])
}

func TestExtractServicesToCheck_WithInvalidFilter(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	// Request with invalid service filter
	req := httptest.NewRequest(http.MethodGet, "/api/playbook-dispatcher/v1/runs?filter[service]=invalid_service", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	result := extractServicesToCheck(ctx, log)

	// Should return empty map for invalid service (fail securely)
	assert.Empty(t, result)
}

func TestExtractServicesToCheck_WithSpaces(t *testing.T) {
	e := echo.New()
	log := zap.NewNop().Sugar()

	// Request with spaces in service filter
	req := httptest.NewRequest(http.MethodGet, "/api/playbook-dispatcher/v1/runs?filter[service]=%20remediations%20", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	result := extractServicesToCheck(ctx, log)

	// Should trim spaces and find the service
	assert.Equal(t, 1, len(result))
	assert.Contains(t, result, "remediations")
}
