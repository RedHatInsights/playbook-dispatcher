package kessel

import (
	"context"
	"testing"

	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCheckPermission_ClientNotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	ctx := context.Background()
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Kessel client not initialized")
	assert.False(t, allowed)
}

func TestCheckPermission_ClientNotInitializedNoIdentity(t *testing.T) {
	// Reset package variables to ensure client is nil
	client = nil
	tokenClient = nil
	rbacClient = nil

	ctx := context.Background() // No identity in context
	log := zap.NewNop().Sugar()

	_, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	// Should error because client is not initialized (fails before identity check)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Kessel client not initialized")
}

func TestCheckPermission_NoIdentityInContext(t *testing.T) {
	// Note: Testing missing identity requires a mock client
	// This test is covered in authorization_mock_test.go
	t.Skip("Covered by TestCheckPermission_WithMock tests - requires functional mock client")
}

func TestCheckPermissionForUpdate_ClientNotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	ctx := context.Background()
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Kessel client not initialized")
	assert.False(t, allowed)
}

func TestExtractUserID_ValidUser(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: identity.User{
				UserID: "test-user-123",
			},
		},
	}

	userID, err := extractUserID(xrhid)

	assert.NoError(t, err)
	assert.Equal(t, "test-user-123", userID)
}

func TestExtractUserID_EmptyUserID(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: identity.User{
				UserID: "",
			},
		},
	}

	userID, err := extractUserID(xrhid)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is empty")
	assert.Empty(t, userID)
}

func TestExtractUserID_UnsupportedType(t *testing.T) {
	tests := []struct {
		name         string
		identityType string
	}{
		{
			name:         "ServiceAccount",
			identityType: "ServiceAccount",
		},
		{
			name:         "System",
			identityType: "System",
		},
		{
			name:         "Unknown",
			identityType: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xrhid := identity.XRHID{
				Identity: identity.Identity{
					Type: tt.identityType,
				},
			}

			userID, err := extractUserID(xrhid)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported identity type")
			assert.Contains(t, err.Error(), tt.identityType)
			assert.Empty(t, userID)
		})
	}
}

func TestGetWorkspaceID_RbacClientNotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	ctx := context.Background()
	log := zap.NewNop().Sugar()

	workspaceID, err := GetWorkspaceID(ctx, "org-123", log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RBAC client not initialized")
	assert.Empty(t, workspaceID)
}

func TestCheckApplicationPermissions_ClientNotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	ctx := context.Background()
	log := zap.NewNop().Sugar()

	allowedApps, err := CheckApplicationPermissions(ctx, "workspace-123", log)

	// Should return error for structural failure (client not initialized)
	// This distinguishes system failures from legitimate "no permissions"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Kessel client not initialized")
	assert.Nil(t, allowedApps)
}

func TestCheckApplicationPermissions_NoIdentityInContext(t *testing.T) {
	// Note: We can't easily test this without a real/mock client
	// because CheckApplicationPermissions calls CheckPermission which requires
	// a functional client. This behavior is covered by the mock tests.
	// This test documents the expected error path.
	t.Skip("Covered by mock tests - requires functional client")
}

func TestCheckApplicationPermissions_InvalidIdentityType(t *testing.T) {
	// Note: We can't easily test this without a real/mock client
	// because CheckApplicationPermissions calls CheckPermission which requires
	// a functional client. This behavior is covered by the mock tests.
	// This test documents the expected error path.
	t.Skip("Covered by mock tests - requires functional client")
}

func TestPrincipalIDFormat(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		expectedID string
	}{
		{
			name:       "simple user id",
			userID:     "user-123",
			expectedID: "redhat/user-123",
		},
		{
			name:       "numeric user id",
			userID:     "12345678",
			expectedID: "redhat/12345678",
		},
		{
			name:       "uuid user id",
			userID:     "550e8400-e29b-41d4-a716-446655440000",
			expectedID: "redhat/550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the format constant produces expected results
			xrhid := identity.XRHID{
				Identity: identity.Identity{
					Type: "User",
					User: identity.User{
						UserID: tt.userID,
					},
				},
			}

			userID, err := extractUserID(xrhid)
			assert.NoError(t, err)

			// Apply the format
			principalID := formatPrincipalID(userID)

			assert.Equal(t, tt.expectedID, principalID)
		})
	}
}

// Helper function to test principal ID formatting
func formatPrincipalID(userID string) string {
	return "redhat/" + userID
}

func TestV2ApplicationPermissions_Coverage(t *testing.T) {
	// Verify that V2ApplicationPermissions map is correctly structured
	assert.NotEmpty(t, V2ApplicationPermissions)
	assert.Equal(t, 3, len(V2ApplicationPermissions))

	// Verify expected applications exist
	assert.Contains(t, V2ApplicationPermissions, "config_manager")
	assert.Contains(t, V2ApplicationPermissions, "remediations")
	assert.Contains(t, V2ApplicationPermissions, "tasks")

	// Verify permissions are correct
	assert.Equal(t, PermissionConfigManagerRunView, V2ApplicationPermissions["config_manager"])
	assert.Equal(t, PermissionRemediationsRunView, V2ApplicationPermissions["remediations"])
	assert.Equal(t, PermissionTasksRunView, V2ApplicationPermissions["tasks"])
}

func TestPlaybookPermissions_Coverage(t *testing.T) {
	// Verify that PlaybookPermissions map is correctly structured
	assert.NotEmpty(t, PlaybookPermissions)
	assert.Equal(t, 2, len(PlaybookPermissions))

	// Verify expected permissions exist
	assert.Contains(t, PlaybookPermissions, PermissionRunRead)
	assert.Contains(t, PlaybookPermissions, PermissionRunWrite)

	// Verify permission details
	readPerm := PlaybookPermissions[PermissionRunRead]
	assert.Equal(t, ResourceTypeWorkspace, readPerm.ResourceType)
	assert.Equal(t, PermissionRunRead, readPerm.Relation)
	assert.NotEmpty(t, readPerm.Description)

	writePerm := PlaybookPermissions[PermissionRunWrite]
	assert.Equal(t, ResourceTypeWorkspace, writePerm.ResourceType)
	assert.Equal(t, PermissionRunWrite, writePerm.Relation)
	assert.NotEmpty(t, writePerm.Description)
}
