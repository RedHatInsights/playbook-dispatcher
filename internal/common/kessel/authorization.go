// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"context"
	"errors"
	"fmt"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// validateClientAndIdentity performs common validation checks for authorization requests
// Returns the XRHID and principal ID if validation succeeds
func validateClientAndIdentity(ctx context.Context) (identity.XRHID, string, error) {
	if client == nil {
		return identity.XRHID{}, "", errors.New("Kessel client not initialized")
	}

	// Extract identity from context
	value := ctx.Value(identity.Key)
	if value == nil {
		return identity.XRHID{}, "", errors.New("no identity found in context")
	}

	xrhid, ok := value.(identity.XRHID)
	if !ok {
		return identity.XRHID{}, "", errors.New("identity in context is not of type XRHID")
	}

	// Extract user ID from identity
	userID, err := extractUserID(xrhid)
	if err != nil {
		return identity.XRHID{}, "", fmt.Errorf("failed to extract user ID: %w", err)
	}

	principalID := fmt.Sprintf(PrincipalIDFormat, userID)

	return xrhid, principalID, nil
}

// buildKesselReferences creates the resource and subject references for Kessel requests
func buildKesselReferences(workspaceID, principalID string) (*kesselv2.ResourceReference, *kesselv2.SubjectReference) {
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

	return object, subject
}

// getAuthCallOptions returns gRPC call options with authentication token if auth is enabled
func getAuthCallOptions() ([]grpc.CallOption, error) {
	var opts []grpc.CallOption
	if tokenClient != nil {
		tokenOpts, err := client.GetTokenCallOption()
		if err != nil {
			return nil, fmt.Errorf("failed to get auth token: %w", err)
		}
		opts = tokenOpts
	}
	return opts, nil
}

// CheckPermission performs a Kessel authorization check for a user's permission on a workspace
//
// Parameters:
//   - ctx: Request context containing identity information
//   - workspaceID: The workspace resource ID to check permission against
//   - permission: The permission/relation to check (e.g., PermissionPlaybookRun)
//   - log: Logger for debugging and error reporting
//
// Returns:
//   - allowed: true if the user has the permission, false otherwise
//   - err: error if the check failed (network error, invalid identity, etc.)
//
// Example usage:
//
//	allowed, err := kessel.CheckPermission(ctx, workspaceID, kessel.PermissionPlaybookRun, log)
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
	object, subject := buildKesselReferences(workspaceID, principalID)

	// Get authentication options
	opts, err := getAuthCallOptions()
	if err != nil {
		return false, err
	}

	// Perform the permission check
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

	response, err := client.KesselInventoryService.Check(ctx, request, opts...)
	if err != nil {
		return false, fmt.Errorf("Kessel check failed: %w", err)
	}

	allowed := response.GetAllowed() == kesselv2.Allowed_ALLOWED_TRUE

	log.Debugw("Kessel permission check complete",
		"workspace_id", workspaceID,
		"permission", permission,
		"principal_id", principalID,
		"allowed", allowed)

	return allowed, nil
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
	object, subject := buildKesselReferences(workspaceID, principalID)

	// Get authentication options
	opts, err := getAuthCallOptions()
	if err != nil {
		return false, err
	}

	// Perform the permission check for update
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

	response, err := client.KesselInventoryService.CheckForUpdate(ctx, request, opts...)
	if err != nil {
		return false, fmt.Errorf("Kessel check for update failed: %w", err)
	}

	allowed := response.GetAllowed() == kesselv2.Allowed_ALLOWED_TRUE

	log.Debugw("Kessel permission check for update complete",
		"workspace_id", workspaceID,
		"permission", permission,
		"principal_id", principalID,
		"allowed", allowed)

	return allowed, nil
}

// extractUserID extracts the user ID from the identity
// For v1 of platform-go-middlewares, we use the User.UserID field
func extractUserID(xrhid identity.XRHID) (string, error) {
	if xrhid.Identity.Type != "User" {
		return "", fmt.Errorf("unsupported identity type: %s (only User is supported in v1)", xrhid.Identity.Type)
	}

	if xrhid.Identity.User.UserID == "" {
		return "", errors.New("user ID is empty")
	}

	return xrhid.Identity.User.UserID, nil
}

// GetWorkspaceID retrieves the default workspace ID for an organization
// This queries the RBAC service for the organization's default workspace
func GetWorkspaceID(ctx context.Context, orgID string, log *zap.SugaredLogger) (string, error) {
	if rbacClient == nil {
		return "", errors.New("RBAC client not initialized")
	}

	log.Debugw("Looking up default workspace ID", "org_id", orgID)

	workspaceID, err := rbacClient.GetDefaultWorkspaceID(ctx, orgID)
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
//         Returns error immediately on first structural failure
//         nil for successful checks (even if user has no permissions)
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
	// Validate client and identity upfront using centralized validation
	// This detects structural failures early (client not initialized, missing identity,
	// wrong identity type, empty UserID) before checking any permissions
	_, _, err := validateClientAndIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot perform authorization checks: %w", err)
	}

	allowedApps := make([]string, 0, len(V2ApplicationPermissions))

	// Loop through each application and check its permission
	for appName, permission := range V2ApplicationPermissions {
		allowed, err := CheckPermission(ctx, workspaceID, permission, log)
		if err != nil {
			// Any error from CheckPermission indicates a structural failure
			// (client, identity, or auth issues) - return immediately
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
