package public

import (
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/api/rbac"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/kessel"
	"playbook-dispatcher/internal/common/unleash/features"
	"playbook-dispatcher/internal/common/utils"
	"reflect"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"go.uber.org/zap"
)

// getAllowedServices extracts allowed services from permissions in the context.
// It is shared between multiple controllers
// and separates rbac and kessel logic from the main controller logic.
// It plugs into the existing access flow and returns a list of allowed services.
//
// This function supports four authorization modes via feature flags:
// - rbac-only: Use RBAC only (default, backward compatible)
// - both-rbac-enforces: Run both RBAC and Kessel, log comparison, enforce RBAC result (validation mode)
// - both-kessel-enforces: Run both RBAC and Kessel, log comparison, enforce Kessel result (transition mode)
// - kessel-only: Use Kessel only (final migration state)
func getAllowedServices(ctx echo.Context) []string {
	cfg := config.Get()
	log := utils.GetLogFromEcho(ctx)

	// Get authorization mode from feature flags (Unleash or environment variable)
	mode := features.GetKesselAuthModeWithContext(ctx.Request().Context(), cfg, log)

	switch mode {
	case config.KesselModeRBACOnly:
		// RBAC only - backward compatible behavior
		log.Debugw("Using RBAC-only authorization mode")
		return getRbacAllowedServices(ctx)

	case config.KesselModeBothRBACEnforces:
		// Validation mode: Run both, compare results, enforce RBAC
		log.Debugw("Using both-rbac-enforces authorization mode (validation)")
		rbacServices := getRbacAllowedServices(ctx)
		kesselServices := getKesselAllowedServices(ctx, log)
		logComparison(rbacServices, kesselServices, log)
		return rbacServices

	case config.KesselModeBothKesselEnforces:
		// Transition mode: Run both, compare results, enforce Kessel
		log.Debugw("Using both-kessel-enforces authorization mode (transition)")
		rbacServices := getRbacAllowedServices(ctx)
		kesselServices := getKesselAllowedServices(ctx, log)
		logComparison(rbacServices, kesselServices, log)
		return kesselServices

	case config.KesselModeKesselOnly:
		// Kessel only - final migration state
		log.Debugw("Using kessel-only authorization mode")
		return getKesselAllowedServices(ctx, log)

	default:
		// Fallback to RBAC if mode is unknown
		log.Warnw("Unknown Kessel authorization mode, falling back to RBAC",
			"mode", mode)
		return getRbacAllowedServices(ctx)
	}
}

// getRbacAllowedServices extracts allowed services from RBAC permissions
func getRbacAllowedServices(ctx echo.Context) []string {
	permissions := middleware.GetPermissions(ctx)
	allowedServices := rbac.GetPredicateValues(permissions, "service")
	return allowedServices
}

// getKesselAllowedServices queries Kessel for allowed services
func getKesselAllowedServices(ctx echo.Context, log *zap.SugaredLogger) []string {
	// Extract identity from context
	xrhid := identity.GetIdentity(ctx.Request().Context())
	orgID := xrhid.Identity.OrgID

	// Get workspace ID for the organization
	workspaceID, err := kessel.GetWorkspaceID(ctx.Request().Context(), orgID, log)
	if err != nil {
		log.Errorw("Failed to get workspace ID", "error", err, "org_id", orgID)
		return []string{} // Return empty list on error
	}

	// Define service permissions to check
	// TODO: Make this configurable via config file or API
	permissions := kessel.ServicePermissions{
		Services: []kessel.ServicePermission{
			{
				Name:       "config_manager",
				Permission: "playbook_dispatcher_config_manager_run_view",
			},
			{
				Name:       "remediations",
				Permission: "playbook_dispatcher_remediations_run_view",
			},
			{
				Name:       "tasks",
				Permission: "playbook_dispatcher_tasks_run_view",
			},
		},
	}

	// Check permissions via Kessel
	allowedServices, err := kessel.CheckPermissions(ctx.Request().Context(), workspaceID, permissions, log)
	if err != nil {
		log.Errorw("Failed to check permissions", "error", err)
		return []string{} // Return empty list on error
	}

	return allowedServices
}

// logComparison compares RBAC and Kessel results and logs any discrepancies
func logComparison(rbacServices, kesselServices []string, log *zap.SugaredLogger) {
	// Sort for comparison
	sortedRbac := make([]string, len(rbacServices))
	copy(sortedRbac, rbacServices)
	sort.Strings(sortedRbac)

	sortedKessel := make([]string, len(kesselServices))
	copy(sortedKessel, kesselServices)
	sort.Strings(sortedKessel)

	if !reflect.DeepEqual(sortedRbac, sortedKessel) {
		log.Warnw("RBAC and Kessel permission mismatch",
			"rbac_services", sortedRbac,
			"kessel_services", sortedKessel)
		// TODO: Increment Prometheus metric for mismatch
	} else {
		log.Debugw("RBAC and Kessel permissions match",
			"services", sortedRbac)
	}
}
