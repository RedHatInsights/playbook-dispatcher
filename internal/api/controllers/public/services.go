package public

import (
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/api/rbac"

	"github.com/labstack/echo/v4"
)

// getAllowedServices extracts allowed services from permissions in the context.
// It is shared between multiple controllers
// and separates rbac and kessel logic from the main controller logic.
// It plugs into the existing access flow and returns a list of allowed services.
func getAllowedServices(ctx echo.Context) []string {
	permissions := middleware.GetPermissions(ctx)
	allowedServices := rbac.GetPredicateValues(permissions, "service")

	return allowedServices
}
