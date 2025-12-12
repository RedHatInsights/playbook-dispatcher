// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

// ServicePermission represents a mapping between a service name and its Kessel permission
type ServicePermission struct {
	Name       string `json:"name"`
	Permission string `json:"permission"`
}

// ServicePermissions holds the collection of service permissions to check
type ServicePermissions struct {
	Services []ServicePermission `json:"services"`
}

const (
	// V1 Permissions - Current RBAC implementation

	// PermissionRunRead grants read access to playbook runs
	// Maps to RBAC permission: playbook-dispatcher:run:read
	PermissionRunRead = "playbook_dispatcher_run_read"

	// PermissionRunWrite grants write access to playbook runs
	// Maps to RBAC permission: playbook-dispatcher:run:write
	PermissionRunWrite = "playbook_dispatcher_run_write"
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
