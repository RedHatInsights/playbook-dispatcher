// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

// Playbook Dispatcher specific permissions for Kessel authorization
// These map to the permissions defined in the RBAC Kessel schema
// See: rbac-config PR #699 - configs/stage/schemas/src/playbook-dispatcher.ksl
//
// V1 Permissions (Current RBAC implementation):
// - playbook-dispatcher:run:read -> playbook_dispatcher_run_read
// - playbook-dispatcher:run:write -> playbook_dispatcher_run_write
//
// V2 Permissions (Kessel workspace-based, feature-flagged):
// - playbook-dispatcher:remediations_run:read -> playbook_dispatcher_remediations_run_view
// - playbook-dispatcher:tasks_run:read -> playbook_dispatcher_tasks_run_view
// - playbook-dispatcher:config_manager_run:read -> playbook_dispatcher_config_manager_run_view

const (
	// V1 Permissions - Current RBAC implementation

	// PermissionRunRead grants read access to playbook runs
	// Maps to RBAC permission: playbook-dispatcher:run:read
	PermissionRunRead = "playbook_dispatcher_run_read"

	// PermissionRunWrite grants write access to playbook runs
	// Maps to RBAC permission: playbook-dispatcher:run:write
	PermissionRunWrite = "playbook_dispatcher_run_write"

	// V2 Permissions - Service-specific Kessel workspace permissions (feature-flagged)

	// PermissionRemediationsRunView grants view access to remediation playbook runs
	// Maps to RBAC permission: playbook-dispatcher:remediations_run:read
	PermissionRemediationsRunView = "playbook_dispatcher_remediations_run_view"

	// PermissionTasksRunView grants view access to task playbook runs
	// Maps to RBAC permission: playbook-dispatcher:tasks_run:read
	PermissionTasksRunView = "playbook_dispatcher_tasks_run_view"

	// PermissionConfigManagerRunView grants view access to config-manager playbook runs
	// Maps to RBAC permission: playbook-dispatcher:config_manager_run:read
	PermissionConfigManagerRunView = "playbook_dispatcher_config_manager_run_view"
)

const (
	// ResourceTypeWorkspace represents a workspace resource in Kessel
	ResourceTypeWorkspace = "workspace"

	// ResourceTypePrincipal represents a user/service account principal in Kessel
	ResourceTypePrincipal = "principal"
)

const (
	// ReporterTypeRBAC indicates the resource is reported by the RBAC service
	ReporterTypeRBAC = "rbac"
)

// Permission represents a Kessel permission check definition
type Permission struct {
	// Resource type (e.g., "workspace")
	ResourceType string

	// The permission/relation being checked (e.g., "playbook_run")
	Relation string

	// Description of what this permission allows
	Description string
}

// PrincipalIDFormat is the format for principal IDs in Kessel
// Example: "redhat/user-123" or "redhat/service-account-456"
const PrincipalIDFormat = "redhat/%s"

// PlaybookPermissions defines V1 RBAC permissions (used by existing authorization code)
// V1 permissions remain unchanged for backwards compatibility
var PlaybookPermissions = map[string]Permission{
	PermissionRunRead: {
		ResourceType: ResourceTypeWorkspace,
		Relation:     PermissionRunRead,
		Description:  "Allows viewing playbook run status, results, and history",
	},
	PermissionRunWrite: {
		ResourceType: ResourceTypeWorkspace,
		Relation:     PermissionRunWrite,
		Description:  "Allows creating new playbook runs and canceling existing runs",
	},
}

// V2ApplicationPermissions maps application names to their Kessel permission names
// Used for checking service-specific access via Kessel workspace permissions
//
// The application names match the "service" field values used in the database:
// - "config_manager" -> service field in runs table
// - "remediations" -> service field in runs table
// - "tasks" -> service field in runs table
var V2ApplicationPermissions = map[string]string{
	"config_manager": PermissionConfigManagerRunView,
	"remediations":   PermissionRemediationsRunView,
	"tasks":          PermissionTasksRunView,
}
