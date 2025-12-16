package middleware

import (
	"net/http"
	"net/http/httptest"
	"playbook-dispatcher/internal/api/rbac"
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
