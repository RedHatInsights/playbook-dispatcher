// Package middleware provides HTTP middleware for the playbook-dispatcher API.
//
// Updated in collaboration with AI
package middleware

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/rbac"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/kessel"
	"playbook-dispatcher/internal/common/unleash/features"
	"playbook-dispatcher/internal/common/utils"
	"reflect"
	"sort"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type permissionsKeyType int
type allowedServicesKeyType int

const permissionsKey permissionsKeyType = iota
const allowedServicesKey allowedServicesKeyType = iota

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
			log := utils.GetLogFromEcho(c)

			// Determine authorization mode (per-request, dynamic)
			mode := features.GetKesselAuthModeWithContext(req.Context(), cfg, log)

			var permissions []rbac.Access

			// TIER 1: RBAC base permission check (skip in kessel-only mode)
			// remains unchanged from RBAC v1 implementation
			if mode != config.KesselModeKesselOnly {
				var err error

				// Log RBAC v1 calls when Kessel is active for comparison/debugging
				if mode == config.KesselModeBothRBACEnforces || mode == config.KesselModeBothKesselEnforces {
					log.Debugw("RBAC v1 permission lookup started for Kessel comparison", "mode", mode)
					start := time.Now()
					permissions, err = client.GetPermissions(req.Context())
					if err == nil {
						allowedServices := getRbacAllowedServices(permissions)
						log.Debugw("RBAC v1 permission lookup succeeded",
							"duration_ms", time.Since(start).Milliseconds(),
							"permission_count", len(permissions),
							"allowed_services", allowedServices,
							"mode", mode)
					} else {
						log.Debugw("RBAC v1 permission lookup failed",
							"duration_ms", time.Since(start).Milliseconds(),
							"error", err,
							"mode", mode)
					}
				} else {
					permissions, err = client.GetPermissions(req.Context())
				}

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
			}

			// TIER 2: Service-level authorization
			allowedServices := computeAllowedServices(c, permissions, mode, log)

			// In Kessel-enforcing modes, empty allowedServices means no permissions (403)
			if len(allowedServices) == 0 {
				switch mode {
				case config.KesselModeBothKesselEnforces, config.KesselModeKesselOnly:
					log.Debugw("User has no Kessel permissions to any services", "mode", mode)
					return echo.NewHTTPError(http.StatusForbidden)
				}
				// In RBAC modes, empty means all services (continue)
			}

			// Cache allowed services for handler
			utils.SetRequestContextValue(c, allowedServicesKey, allowedServices)

			return next(c)
		}
	}
}

func GetPermissions(c echo.Context) []rbac.Access {
	value := c.Request().Context().Value(permissionsKey)
	if value == nil {
		return []rbac.Access{}
	}
	permissions, ok := value.([]rbac.Access)
	if !ok {
		return []rbac.Access{}
	}
	return permissions
}

// GetAllowedServices retrieves the cached allowed services from context
func GetAllowedServices(c echo.Context) []string {
	value := c.Request().Context().Value(allowedServicesKey)
	if value == nil {
		return []string{}
	}
	services, ok := value.([]string)
	if !ok {
		return []string{}
	}
	return services
}

// computeAllowedServices determines which services the user can access
// based on the authorization mode
func computeAllowedServices(ctx echo.Context, rbacPermissions []rbac.Access, mode string, log *zap.SugaredLogger) []string {
	switch mode {
	case config.KesselModeRBACOnly:
		log.Debugw("Using RBAC-only authorization mode")
		return getRbacAllowedServices(rbacPermissions)

	case config.KesselModeBothRBACEnforces:
		log.Debugw("Using both-rbac-enforces authorization mode (validation)")
		rbacServices := getRbacAllowedServices(rbacPermissions)
		kesselServices := getKesselAllowedServices(ctx, log)
		logComparison(ctx, rbacServices, kesselServices, log)
		return rbacServices

	case config.KesselModeBothKesselEnforces:
		log.Debugw("Using both-kessel-enforces authorization mode (transition)")
		rbacServices := getRbacAllowedServices(rbacPermissions)
		kesselServices := getKesselAllowedServices(ctx, log)
		logComparison(ctx, rbacServices, kesselServices, log)
		return kesselServices

	case config.KesselModeKesselOnly:
		log.Debugw("Using kessel-only authorization mode")
		return getKesselAllowedServices(ctx, log)

	default:
		log.Warnw("Unknown Kessel authorization mode, falling back to RBAC",
			"mode", mode)
		return getRbacAllowedServices(rbacPermissions)
	}
}

// getRbacAllowedServices extracts allowed services from RBAC permissions
func getRbacAllowedServices(permissions []rbac.Access) []string {
	return rbac.GetPredicateValues(permissions, "service")
}

// getKesselAllowedServices queries Kessel for allowed services
func getKesselAllowedServices(ctx echo.Context, log *zap.SugaredLogger) []string {
	// Extract identity from context
	xrhid := identity.GetIdentity(ctx.Request().Context())
	orgID := xrhid.Identity.OrgID
	identityType := xrhid.Identity.Type

	// Provide fallback for missing identity fields
	if orgID == "" {
		orgID = "unknown"
	}
	if identityType == "" {
		identityType = "unknown"
	}

	// Extract user ID based on identity type with nil guards
	var userID string
	switch identityType {
	case "User":
		if xrhid.Identity.User != nil {
			userID = xrhid.Identity.User.UserID
		}
		if userID == "" {
			userID = "unknown"
		}
	case "ServiceAccount":
		if xrhid.Identity.ServiceAccount != nil {
			userID = xrhid.Identity.ServiceAccount.UserId
		}
		if userID == "" {
			userID = "unknown"
		}
	default:
		userID = "unknown"
	}

	// Get workspace ID for the organization
	workspaceID, err := kessel.GetWorkspaceID(ctx.Request().Context(), orgID, log)
	if err != nil {
		log.Errorw("Kessel authorization error",
			"error", err,
			"org_id", orgID,
			"identity_type", identityType,
			"user_id", userID)
		instrumentation.KesselAuthorizationError(ctx)
		return []string{} // Return empty list on error
	}

	// Check permissions via Kessel (uses V2ApplicationPermissions map)
	allowedServices, err := kessel.CheckApplicationPermissions(ctx.Request().Context(), workspaceID, log)
	if err != nil {
		log.Errorw("Kessel authorization error",
			"error", err,
			"org_id", orgID,
			"workspace_id", workspaceID,
			"identity_type", identityType,
			"user_id", userID)
		instrumentation.KesselAuthorizationError(ctx)
		return []string{} // Return empty list on error
	}

	if len(allowedServices) == 0 {
		log.Debugw("Kessel authorization returned no services",
			"org_id", orgID,
			"workspace_id", workspaceID,
			"identity_type", identityType,
			"user_id", userID)
		instrumentation.KesselAuthorizationFailed(ctx)
	} else {
		log.Debugw("Kessel authorization succeeded",
			"org_id", orgID,
			"workspace_id", workspaceID,
			"identity_type", identityType,
			"user_id", userID,
			"allowed_services", allowedServices)
		instrumentation.KesselAuthorizationPassed(ctx)
	}

	return allowedServices
}

// logComparison compares RBAC and Kessel results and logs any discrepancies
func logComparison(ctx echo.Context, rbacServices, kesselServices []string, log *zap.SugaredLogger) {
	// Sort for comparison
	sortedRbac := make([]string, len(rbacServices))
	copy(sortedRbac, rbacServices)
	sort.Strings(sortedRbac)

	sortedKessel := make([]string, len(kesselServices))
	copy(sortedKessel, kesselServices)
	sort.Strings(sortedKessel)

	// Extract identity information for diagnostics (used by both match and mismatch)
	// Guard against malformed identities
	xrhid := identity.GetIdentity(ctx.Request().Context())
	orgID := xrhid.Identity.OrgID
	identityType := xrhid.Identity.Type

	// Provide fallback for missing org ID
	if orgID == "" {
		orgID = "unknown"
	}
	if identityType == "" {
		identityType = "unknown"
	}

	// Extract user ID based on identity type
	var userID string
	switch identityType {
	case "User":
		if xrhid.Identity.User != nil {
			userID = xrhid.Identity.User.UserID
		}
		if userID == "" {
			userID = "unknown"
		}
	case "ServiceAccount":
		if xrhid.Identity.ServiceAccount != nil {
			userID = xrhid.Identity.ServiceAccount.UserId
		}
		if userID == "" {
			userID = "unknown"
		}
	default:
		userID = "unknown"
	}

	if !reflect.DeepEqual(sortedRbac, sortedKessel) {
		log.Warnw("RBAC and Kessel permission mismatch",
			"rbac_services", sortedRbac,
			"kessel_services", sortedKessel,
			"org_id", orgID,
			"identity_type", identityType,
			"user_id", userID)
		instrumentation.KesselRbacMismatch(ctx)
	} else {
		log.Debugw("RBAC and Kessel permissions match",
			"rbac_services", sortedRbac,
			"kessel_services", sortedKessel,
			"org_id", orgID,
			"identity_type", identityType,
			"user_id", userID)
		instrumentation.KesselRbacMatch(ctx)
	}
}
