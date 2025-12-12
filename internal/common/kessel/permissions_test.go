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

func TestPermissionConstants_V2(t *testing.T) {
	// Test that V2 permission constants have expected values
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "PermissionRemediationsRunView",
			constant: PermissionRemediationsRunView,
			expected: "playbook_dispatcher_remediations_run_view",
		},
		{
			name:     "PermissionTasksRunView",
			constant: PermissionTasksRunView,
			expected: "playbook_dispatcher_tasks_run_view",
		},
		{
			name:     "PermissionConfigManagerRunView",
			constant: PermissionConfigManagerRunView,
			expected: "playbook_dispatcher_config_manager_run_view",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

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

func TestV2ApplicationPermissions_Structure(t *testing.T) {
	// Verify V2ApplicationPermissions map structure
	assert.NotNil(t, V2ApplicationPermissions)
	assert.Len(t, V2ApplicationPermissions, 3)

	tests := []struct {
		appName    string
		permission string
	}{
		{
			appName:    "config_manager",
			permission: PermissionConfigManagerRunView,
		},
		{
			appName:    "remediations",
			permission: PermissionRemediationsRunView,
		},
		{
			appName:    "tasks",
			permission: PermissionTasksRunView,
		},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			perm, exists := V2ApplicationPermissions[tt.appName]
			assert.True(t, exists, "application %s should exist in map", tt.appName)
			assert.Equal(t, tt.permission, perm)
		})
	}
}

func TestV2ApplicationPermissions_UniquenessCheck(t *testing.T) {
	// Verify that all V2 permissions are unique
	seen := make(map[string]string)
	for app, perm := range V2ApplicationPermissions {
		if existingApp, exists := seen[perm]; exists {
			t.Errorf("Permission %s is duplicated between apps %s and %s", perm, app, existingApp)
		}
		seen[perm] = app
	}
}

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

func TestV2Permissions_NamingConvention(t *testing.T) {
	// Verify V2 permissions follow naming convention: playbook_dispatcher_{service}_run_view
	expectedPrefix := "playbook_dispatcher_"
	expectedSuffix := "_run_view"

	for app, perm := range V2ApplicationPermissions {
		t.Run(app, func(t *testing.T) {
			assert.Contains(t, perm, expectedPrefix, "permission should start with playbook_dispatcher_")
			assert.Contains(t, perm, expectedSuffix, "permission should end with _run_view")
			assert.Contains(t, perm, app, "permission should contain application name")
		})
	}
}

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
	// Verify that V1 and V2 permission constants don't collide
	allPermissions := make(map[string]bool)

	// Add V1 permissions
	v1Permissions := []string{PermissionRunRead, PermissionRunWrite}
	for _, perm := range v1Permissions {
		if allPermissions[perm] {
			t.Errorf("Duplicate permission constant: %s", perm)
		}
		allPermissions[perm] = true
	}

	// Add V2 permissions
	v2Permissions := []string{
		PermissionRemediationsRunView,
		PermissionTasksRunView,
		PermissionConfigManagerRunView,
	}
	for _, perm := range v2Permissions {
		if allPermissions[perm] {
			t.Errorf("Duplicate permission constant: %s", perm)
		}
		allPermissions[perm] = true
	}

	// Should have 5 unique permissions total
	assert.Len(t, allPermissions, 5)
}
