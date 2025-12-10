// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"context"
	"errors"
	"fmt"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// validateClientAndIdentity performs common validation checks for authorization requests
// Returns the XRHID and principal ID if validation succeeds
func validateClientAndIdentity(ctx context.Context) (identity.XRHID, string, error) {
	if globalManager == nil || globalManager.client == nil {
		return identity.XRHID{}, "", errors.New("Kessel client not initialized")
	}

	// Extract identity from context using v2 middleware
	xrhid := identity.GetIdentity(ctx)

	// Extract user ID from identity
	userID, err := extractUserID(xrhid)
	if err != nil {
		return identity.XRHID{}, "", fmt.Errorf("failed to extract user ID: %w", err)
	}

	principalID := fmt.Sprintf(PrincipalIDFormat, userID)

	return xrhid, principalID, nil
}

// buildKesselReferences creates the resource and subject references for Kessel requests
func buildKesselReferences(workspaceID, principalID string) (*kesselv2.ResourceReference, *kesselv2.SubjectReference, error) {
	// Validate inputs
	if workspaceID == "" {
		return nil, nil, errors.New("workspaceID cannot be empty")
	}
	if principalID == "" {
		return nil, nil, errors.New("principalID cannot be empty")
	}

	// Build resource reference (workspace)
	object := &kesselv2.ResourceReference{
		ResourceType: ResourceTypeWorkspace,
		ResourceId:   workspaceID,
		Reporter: &kesselv2.ReporterReference{
			Type: ReporterTypeRBAC,
		},
	}

	// Build subject reference (user/service account)
	subject := &kesselv2.SubjectReference{
		Resource: &kesselv2.ResourceReference{
			ResourceType: ResourceTypePrincipal,
			ResourceId:   principalID,
			Reporter: &kesselv2.ReporterReference{
				Type: ReporterTypeRBAC,
			},
		},
	}

	return object, subject, nil
}

// getAuthCallOptions returns gRPC call options with authentication token if auth is enabled
func getAuthCallOptions() ([]grpc.CallOption, error) {
	var opts []grpc.CallOption
	// Check tokenClient to determine if auth is enabled; the client.GetTokenCallOption()
	// method internally manages token retrieval and caching
	if globalManager != nil && globalManager.tokenClient != nil {
		tokenOpts, err := globalManager.client.GetTokenCallOption()
		if err != nil {
			return nil, fmt.Errorf("failed to get auth token: %w", err)
		}
		opts = tokenOpts
	}
	return opts, nil
}

// checkPermissionInternal is the shared internal helper for permission checks
// This reduces duplication between CheckPermission and CheckPermissionForUpdate
func checkPermissionInternal(
	ctx context.Context,
	workspaceID string,
	permission string,
	log *zap.SugaredLogger,
	xrhid identity.XRHID,
	principalID string,
	object *kesselv2.ResourceReference,
	subject *kesselv2.SubjectReference,
	opts []grpc.CallOption,
	useCheckForUpdate bool,
) (bool, error) {
	var allowed bool

	if useCheckForUpdate {
		request := &kesselv2.CheckForUpdateRequest{
			Object:   object,
			Relation: permission,
			Subject:  subject,
		}

		log.Debugw("Performing Kessel permission check for update",
			"workspace_id", workspaceID,
			"permission", permission,
			"principal_id", principalID,
			"org_id", xrhid.Identity.OrgID)

		response, err := globalManager.client.KesselInventoryService.CheckForUpdate(ctx, request, opts...)
		if err != nil {
			return false, fmt.Errorf("Kessel check for update failed: %w", err)
		}
		allowed = response.GetAllowed() == kesselv2.Allowed_ALLOWED_TRUE

		log.Debugw("Kessel permission check for update complete",
			"workspace_id", workspaceID,
			"permission", permission,
			"principal_id", principalID,
			"allowed", allowed)
	} else {
		request := &kesselv2.CheckRequest{
			Object:   object,
			Relation: permission,
			Subject:  subject,
		}

		log.Debugw("Performing Kessel permission check",
			"workspace_id", workspaceID,
			"permission", permission,
			"principal_id", principalID,
			"org_id", xrhid.Identity.OrgID)

		response, err := globalManager.client.KesselInventoryService.Check(ctx, request, opts...)
		if err != nil {
			return false, fmt.Errorf("Kessel check failed: %w", err)
		}
		allowed = response.GetAllowed() == kesselv2.Allowed_ALLOWED_TRUE

		log.Debugw("Kessel permission check complete",
			"workspace_id", workspaceID,
			"permission", permission,
			"principal_id", principalID,
			"allowed", allowed)
	}

	return allowed, nil
}

// CheckPermission performs a Kessel authorization check for a user's permission on a workspace
//
// Parameters:
//   - ctx: Request context containing identity information
//   - workspaceID: The workspace resource ID to check permission against
//   - permission: The permission/relation to check (e.g., PermissionRemediationsRunView)
//   - log: Logger for debugging and error reporting
//
// Returns:
//   - allowed: true if the user has the permission, false otherwise
//   - err: error if the check failed (network error, invalid identity, etc.)
//
// Example usage:
//
//	allowed, err := kessel.CheckPermission(ctx, workspaceID, kessel.PermissionRemediationsRunView, log)
//	if err != nil {
//	    log.Error("Authorization check failed", err)
//	    return http.StatusInternalServerError
//	}
//	if !allowed {
//	    return http.StatusForbidden
//	}
func CheckPermission(ctx context.Context, workspaceID string, permission string, log *zap.SugaredLogger) (bool, error) {
	// Validate client and extract identity
	xrhid, principalID, err := validateClientAndIdentity(ctx)
	if err != nil {
		return false, err
	}

	// Build Kessel request references
	object, subject, err := buildKesselReferences(workspaceID, principalID)
	if err != nil {
		return false, err
	}

	// Get authentication options
	opts, err := getAuthCallOptions()
	if err != nil {
		return false, err
	}

	// Use shared helper for the actual check
	return checkPermissionInternal(ctx, workspaceID, permission, log, xrhid, principalID, object, subject, opts, false)
}

// CheckPermissionForUpdate performs a Kessel authorization check using CheckForUpdate
// This is used for update operations where Kessel needs to verify the current state
//
// Parameters are the same as CheckPermission
func CheckPermissionForUpdate(ctx context.Context, workspaceID string, permission string, log *zap.SugaredLogger) (bool, error) {
	// Validate client and extract identity
	xrhid, principalID, err := validateClientAndIdentity(ctx)
	if err != nil {
		return false, err
	}

	// Build Kessel request references
	object, subject, err := buildKesselReferences(workspaceID, principalID)
	if err != nil {
		return false, err
	}

	// Get authentication options
	opts, err := getAuthCallOptions()
	if err != nil {
		return false, err
	}

	// Use shared helper for the actual check
	return checkPermissionInternal(ctx, workspaceID, permission, log, xrhid, principalID, object, subject, opts, true)
}

// extractUserID extracts the user ID from the identity
// Supports both User and ServiceAccount identity types (platform-go-middlewares v2)
func extractUserID(xrhid identity.XRHID) (string, error) {
	switch xrhid.Identity.Type {
	case "User":
		if xrhid.Identity.User == nil || xrhid.Identity.User.UserID == "" {
			return "", errors.New("user ID is empty")
		}
		return xrhid.Identity.User.UserID, nil
	case "ServiceAccount":
		if xrhid.Identity.ServiceAccount == nil || xrhid.Identity.ServiceAccount.UserId == "" {
			return "", errors.New("service account user ID is empty")
		}
		// Note: ServiceAccount uses UserId (lowercase 'd') due to upstream library inconsistency
		return xrhid.Identity.ServiceAccount.UserId, nil
	default:
		return "", fmt.Errorf("unsupported identity type: %s (only User and ServiceAccount are supported)", xrhid.Identity.Type)
	}
}

// GetWorkspaceID retrieves the default workspace ID for an organization
// This queries the RBAC service for the organization's default workspace
func GetWorkspaceID(ctx context.Context, orgID string, log *zap.SugaredLogger) (string, error) {
	if globalManager == nil || globalManager.rbacClient == nil {
		return "", errors.New("RBAC client not initialized")
	}

	log.Debugw("Looking up default workspace ID", "org_id", orgID)

	workspaceID, err := globalManager.rbacClient.GetDefaultWorkspaceID(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to get default workspace ID: %w", err)
	}

	log.Debugw("Found default workspace ID",
		"org_id", orgID,
		"workspace_id", workspaceID)

	return workspaceID, nil
}

// CheckApplicationPermissions checks V2 Kessel permissions for multiple applications
// and returns a list of application names that the user has access to.
//
// This function loops through the V2 application-specific permissions:
// - playbook-dispatcher:config_manager_run:read -> playbook_dispatcher_config_manager_run_view
// - playbook-dispatcher:remediations_run:read -> playbook_dispatcher_remediations_run_view
// - playbook-dispatcher:tasks_run:read -> playbook_dispatcher_tasks_run_view
//
// Parameters:
//   - ctx: Request context containing identity information
//   - workspaceID: The workspace resource ID to check permissions against
//   - log: Logger for debugging and error reporting
//
// Returns:
//   - allowedApps: List of application names the user has access to (e.g., ["remediations", "tasks"])
//   - err: Non-nil error for structural failures (client not initialized, bad config, identity issues)
//     Returns error immediately on first structural failure
//     nil for successful checks (even if user has no permissions)
//
// Error Handling:
//   - Structural failures (client == nil, auth config issues, bad identity): Returns error immediately
//   - This allows callers to distinguish system failures from legitimate authorization denials
//
// Example usage:
//
//	allowedApps, err := kessel.CheckApplicationPermissions(ctx, workspaceID, log)
//	if err != nil {
//	    log.Error("Kessel authorization system failure", err)
//	    return http.StatusServiceUnavailable  // System issue, not auth denial
//	}
//	if len(allowedApps) == 0 {
//	    log.Info("User has no application permissions")
//	    return http.StatusForbidden  // Legitimate authorization denial
//	}
//	// allowedApps might be: []string{"remediations", "config_manager"}
func CheckApplicationPermissions(ctx context.Context, workspaceID string, log *zap.SugaredLogger) ([]string, error) {
	// Validate client and extract identity once (shared across all permission checks)
	// This detects structural failures early and avoids re-extracting identity for each app
	xrhid, principalID, err := validateClientAndIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot perform authorization checks: %w", err)
	}

	// Build Kessel request references once (shared across all permission checks)
	object, subject, err := buildKesselReferences(workspaceID, principalID)
	if err != nil {
		return nil, fmt.Errorf("failed to build Kessel references: %w", err)
	}

	// Get authentication options once (shared across all permission checks)
	opts, err := getAuthCallOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth options: %w", err)
	}

	allowedApps := make([]string, 0, len(V2ApplicationPermissions))

	// Loop through each application and check its permission
	// NOTE: We call checkPermissionInternal directly (instead of CheckPermission) to reuse
	// the resolved identity, principal ID, and Kessel references across all permission checks.
	// This avoids redundant identity extraction and reference building for each application,
	// which is important when checking multiple permissions for the same user.
	for appName, permission := range V2ApplicationPermissions {
		allowed, err := checkPermissionInternal(ctx, workspaceID, permission, log, xrhid, principalID, object, subject, opts, false)
		if err != nil {
			// Any error from checkPermissionInternal indicates a structural failure
			// (network error, auth issues) - return immediately
			return nil, fmt.Errorf("structural failure checking permission for %s: %w", appName, err)
		}

		if allowed {
			allowedApps = append(allowedApps, appName)
			log.Debugw("User has access to application",
				"app", appName,
				"permission", permission)
		} else {
			log.Debugw("User does not have access to application",
				"app", appName,
				"permission", permission)
		}
	}

	log.Infow("Application permission check complete",
		"allowed_apps", allowedApps,
		"total_checked", len(V2ApplicationPermissions))

	return allowedApps, nil
}
