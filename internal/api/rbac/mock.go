package rbac

import (
	"context"
)

type mockImpl struct{}

func (*mockImpl) GetPermissions(ctx context.Context) ([]Access, error) {
	return []Access{
		{
			Permission: "playbook-dispatcher:run:read",
			ResourceDefinitions: []ResourceDefinition{
				{
					AttributeFilter: ResourceDefinitionFilter{[]byte(`{"key": "service", "operation": "equal", "value": ["remediations"]}`)},
				},
			},
		},
		{
			Permission: "playbook-dispatcher:run:read",
			ResourceDefinitions: []ResourceDefinition{
				{
					AttributeFilter: ResourceDefinitionFilter{[]byte(`{"key": "service", "operation": "equal", "value": ["config_manager"]}`)},
				},
			},
		},
		{
			Permission: "playbook-dispatcher:run:read",
			ResourceDefinitions: []ResourceDefinition{
				{
					AttributeFilter: ResourceDefinitionFilter{[]byte(`{"key": "service", "operation": "equal", "value": "test"}`)},
				},
			},
		},
	}, nil
}

func NewMockRbacClient() RbacClient {
	return &mockImpl{}
}
