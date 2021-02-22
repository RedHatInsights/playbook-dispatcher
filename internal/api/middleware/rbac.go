package middleware

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/rbac"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

type permissionsKeyType int

const permissionsKey permissionsKeyType = iota

func EnforcePermissions(cfg *viper.Viper, requiredPermissions ...rbac.RequiredPermission) echo.MiddlewareFunc {
	var client rbac.RbacClient

	if cfg.GetString("rbac.impl") == "impl" {
		client = rbac.NewRbacClient(cfg)
	} else {
		client = rbac.NewMockRbacClient()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			permissions, err := client.GetPermissions(req.Context())
			if err != nil {
				instrumentation.RbacError(c, err)
				return echo.NewHTTPError(http.StatusServiceUnavailable, "error getting permissions from RBAC")
			}

			for _, requiredPermission := range requiredPermissions {
				matchingPermissions := rbac.FilterPermissions(permissions, requiredPermission)

				if len(matchingPermissions) == 0 {
					instrumentation.RbacRejected(c)
					return echo.NewHTTPError(http.StatusForbidden)
				}
			}

			utils.SetRequestContextValue(c, permissionsKey, permissions)

			return next(c)
		}
	}
}

func GetPermissions(c echo.Context) []rbac.Access {
	return c.Request().Context().Value(permissionsKey).([]rbac.Access)
}
