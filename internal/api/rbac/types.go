package rbac

import "context"

type RbacClient interface {
	GetPermissions(ctx context.Context) ([]Access, error)
}

type RequiredPermission struct {
	Application  string
	ResourceType string
	Verb         string
}

func DispatcherPermission(resourceType, verb string) RequiredPermission {
	return RequiredPermission{
		Application:  "playbook-dispatcher",
		ResourceType: resourceType,
		Verb:         verb,
	}
}
