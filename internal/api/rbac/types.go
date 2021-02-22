package rbac

import "context"

type RbacClient interface {
	GetPermissions(ctx context.Context) ([]Access, error)
}

type RequiredPermission struct {
	ResourceType string
	Verb         string
}
