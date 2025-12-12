package kessel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionConstants_V1(t *testing.T) {
	// Test that V1 permission constants have expected values
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "PermissionRunRead",
			constant: PermissionRunRead,
			expected: "playbook_dispatcher_run_read",
		},
		{
			name:     "PermissionRunWrite",
			constant: PermissionRunWrite,
			expected: "playbook_dispatcher_run_write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

// V2 permission constants have been removed - permissions are now defined
// in the ServicePermissions struct by the caller

func TestResourceTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "ResourceTypeWorkspace",
			constant: ResourceTypeWorkspace,
			expected: "workspace",
		},
		{
			name:     "ResourceTypePrincipal",
			constant: ResourceTypePrincipal,
			expected: "principal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestReporterTypeConstants(t *testing.T) {
	assert.Equal(t, "rbac", ReporterTypeRBAC)
}

func TestPrincipalIDFormatConstant(t *testing.T) {
	assert.Equal(t, "redhat/%s", PrincipalIDFormat)
}

func TestPlaybookPermissions_Structure(t *testing.T) {
	// Verify PlaybookPermissions map structure
	assert.NotNil(t, PlaybookPermissions)
	assert.Len(t, PlaybookPermissions, 2)

	// Test read permission
	readPerm, exists := PlaybookPermissions[PermissionRunRead]
	assert.True(t, exists)
	assert.Equal(t, ResourceTypeWorkspace, readPerm.ResourceType)
	assert.Equal(t, PermissionRunRead, readPerm.Relation)
	assert.Contains(t, readPerm.Description, "viewing")
	assert.NotEmpty(t, readPerm.Description)

	// Test write permission
	writePerm, exists := PlaybookPermissions[PermissionRunWrite]
	assert.True(t, exists)
	assert.Equal(t, ResourceTypeWorkspace, writePerm.ResourceType)
	assert.Equal(t, PermissionRunWrite, writePerm.Relation)
	assert.Contains(t, writePerm.Description, "creating")
	assert.NotEmpty(t, writePerm.Description)
}

// V2ApplicationPermissions map has been removed - permissions are now passed
// via ServicePermissions struct by the caller

func TestPermissionType_FieldsPresent(t *testing.T) {
	// Test that the Permission type has all expected fields
	perm := Permission{
		ResourceType: "test-resource",
		Relation:     "test-relation",
		Description:  "test description",
	}

	assert.Equal(t, "test-resource", perm.ResourceType)
	assert.Equal(t, "test-relation", perm.Relation)
	assert.Equal(t, "test description", perm.Description)
}

func TestPlaybookPermissions_Immutability(t *testing.T) {
	// Verify that modifying the map doesn't affect the original
	originalLen := len(PlaybookPermissions)

	// Create a copy and modify it
	testMap := make(map[string]Permission)
	for k, v := range PlaybookPermissions {
		testMap[k] = v
	}
	testMap["test_permission"] = Permission{
		ResourceType: "test",
		Relation:     "test",
		Description:  "test",
	}

	// Original should be unchanged
	assert.Len(t, PlaybookPermissions, originalLen)
	assert.NotContains(t, PlaybookPermissions, "test_permission")
}

// V2 permissions naming convention tests removed - permissions are now
// defined by callers in ServicePermissions struct

func TestV1Permissions_NamingConvention(t *testing.T) {
	// Verify V1 permissions follow naming convention: playbook_dispatcher_run_{action}
	expectedPrefix := "playbook_dispatcher_run_"

	for permName, perm := range PlaybookPermissions {
		t.Run(permName, func(t *testing.T) {
			assert.Contains(t, perm.Relation, expectedPrefix, "permission should start with playbook_dispatcher_run_")
		})
	}
}

func TestPermissionConstants_NoCollisions(t *testing.T) {
	// Verify that V1 permission constants are unique
	allPermissions := make(map[string]bool)

	// Add V1 permissions
	v1Permissions := []string{PermissionRunRead, PermissionRunWrite}
	for _, perm := range v1Permissions {
		if allPermissions[perm] {
			t.Errorf("Duplicate permission constant: %s", perm)
		}
		allPermissions[perm] = true
	}

	// V2 permissions are now defined dynamically by callers

	// Should have 2 V1 permissions
	assert.Len(t, allPermissions, 2)
}
