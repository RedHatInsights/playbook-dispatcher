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
					AttributeFilter: ResourceDefinitionFilter{
						Key:       "service",
						Operation: "equal",
						Value:     ResourceDefinitionFilter_Value{[]byte(`["remediations"]`)},
					},
				},
			},
		},
		{
			Permission: "playbook-dispatcher:run:read",
			ResourceDefinitions: []ResourceDefinition{
				{
					AttributeFilter: ResourceDefinitionFilter{
						Key:       "service",
						Operation: "equal",
						Value:     ResourceDefinitionFilter_Value{[]byte(`["config_manager"]`)},
					},
				},
			},
		},
		{
			Permission: "playbook-dispatcher:run:read",
			ResourceDefinitions: []ResourceDefinition{
				{
					AttributeFilter: ResourceDefinitionFilter{
						Key:       "service",
						Operation: "equal",
						Value:     ResourceDefinitionFilter_Value{[]byte(`"test"`)},
					},
				},
			},
		},
	}, nil
}

func NewMockRbacClient() RbacClient {
	return &mockImpl{}
}
