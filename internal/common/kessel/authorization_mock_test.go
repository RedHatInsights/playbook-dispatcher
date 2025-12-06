package kessel

import (
	"context"
	"errors"
	"testing"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Mock types for testing

// mockKesselInventoryService mocks the Kessel inventory service for testing
type mockKesselInventoryService struct {
	checkFunc          func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error)
	checkForUpdateFunc func(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error)
}

func (m *mockKesselInventoryService) Check(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
	if m.checkFunc != nil {
		return m.checkFunc(ctx, req, opts...)
	}
	// Default: return allowed
	return &kesselv2.CheckResponse{
		Allowed: kesselv2.Allowed_ALLOWED_TRUE,
	}, nil
}

func (m *mockKesselInventoryService) CheckForUpdate(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
	if m.checkForUpdateFunc != nil {
		return m.checkForUpdateFunc(ctx, req, opts...)
	}
	// Default: return allowed
	return &kesselv2.CheckForUpdateResponse{
		Allowed: kesselv2.Allowed_ALLOWED_TRUE,
	}, nil
}

// Stub implementations for unused interface methods
func (m *mockKesselInventoryService) ReportResource(ctx context.Context, req *kesselv2.ReportResourceRequest, opts ...grpc.CallOption) (*kesselv2.ReportResourceResponse, error) {
	return nil, nil
}

func (m *mockKesselInventoryService) DeleteResource(ctx context.Context, req *kesselv2.DeleteResourceRequest, opts ...grpc.CallOption) (*kesselv2.DeleteResourceResponse, error) {
	return nil, nil
}

func (m *mockKesselInventoryService) StreamedListObjects(ctx context.Context, req *kesselv2.StreamedListObjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[kesselv2.StreamedListObjectsResponse], error) {
	return nil, nil
}

// Helper function to create a mock client for testing
// This creates a fake v1beta2.InventoryClient with our mock service
func newMockClient(mockService *mockKesselInventoryService) *v1beta2.InventoryClient {
	// We can't easily mock the struct, so we'll use an unsafe pointer cast
	// This works because in tests we only care about KesselInventoryService field
	// and GetTokenCallOption method
	mock := &v1beta2.InventoryClient{
		KesselInventoryService: mockService,
	}
	return mock
}

// Tests with mocked Kessel client

func TestCheckPermission_WithMock_Allowed(t *testing.T) {
	// Set up mock
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			// Verify request is built correctly
			assert.Equal(t, "workspace-123", req.Object.ResourceId)
			assert.Equal(t, ResourceTypeWorkspace, req.Object.ResourceType)
			assert.Equal(t, ReporterTypeRBAC, req.Object.Reporter.Type)
			assert.Equal(t, PermissionRunRead, req.Relation)
			assert.Equal(t, ResourceTypePrincipal, req.Subject.Resource.ResourceType)
			assert.Equal(t, "redhat/test-user-456", req.Subject.Resource.ResourceId)

			// Return allowed
			return &kesselv2.CheckResponse{
				Allowed: kesselv2.Allowed_ALLOWED_TRUE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user-456"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	// Assert
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermission_WithMock_Denied(t *testing.T) {
	// Set up mock to deny permission
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			return &kesselv2.CheckResponse{
				Allowed: kesselv2.Allowed_ALLOWED_FALSE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunWrite, log)

	// Assert
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermission_WithMock_KesselError(t *testing.T) {
	// Set up mock to return error
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			return nil, errors.New("kessel service unavailable")
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Kessel check failed")
	assert.Contains(t, err.Error(), "kessel service unavailable")
	assert.False(t, allowed)
}

func TestCheckPermission_WithMock_WithIdentityInContext(t *testing.T) {
	// Set up mock
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			return &kesselv2.CheckResponse{
				Allowed: kesselv2.Allowed_ALLOWED_TRUE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Test with identity in context
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "user-789"},
			OrgID: "org-456",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-456", PermissionRunRead, log)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermission_WithMock_NoIdentityInContext(t *testing.T) {
	// Set up mock client (won't be called because identity check happens first)
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			// This should not be called
			t.Fatal("Check should not be called when identity is missing")
			return nil, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context without identity
	ctx := context.Background()
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	// Should fail with "no identity found in context"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no identity found in context")
	assert.False(t, allowed)
}

func TestCheckPermission_WithMock_InvalidIdentityType(t *testing.T) {
	// Set up mock client (won't be called because identity validation happens first)
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			// This should not be called
			t.Fatal("Check should not be called when identity type is invalid")
			return nil, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context with wrong identity type
	ctx := context.WithValue(context.Background(), identity.Key, "not-an-xrhid-struct")
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	// Should fail with "identity in context is not of type XRHID"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity in context is not of type XRHID")
	assert.False(t, allowed)
}

func TestCheckPermission_WithMock_UnsupportedIdentityType(t *testing.T) {
	// Set up mock client (won't be called because user extraction fails first)
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			// This should not be called
			t.Fatal("Check should not be called when identity type is unsupported")
			return nil, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context with unsupported identity type (ServiceAccount instead of User)
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "ServiceAccount",
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	// Should fail with "unsupported identity type"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported identity type")
	assert.Contains(t, err.Error(), "ServiceAccount")
	assert.False(t, allowed)
}

func TestCheckPermissionForUpdate_WithMock_Allowed(t *testing.T) {
	// Set up mock
	mockService := &mockKesselInventoryService{
		checkForUpdateFunc: func(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
			// Verify request is built correctly
			assert.Equal(t, "workspace-789", req.Object.ResourceId)
			assert.Equal(t, ResourceTypeWorkspace, req.Object.ResourceType)
			assert.Equal(t, PermissionRunWrite, req.Relation)
			assert.Contains(t, req.Subject.Resource.ResourceId, "redhat/test-user")

			return &kesselv2.CheckForUpdateResponse{
				Allowed: kesselv2.Allowed_ALLOWED_TRUE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowed, err := CheckPermissionForUpdate(ctx, "workspace-789", PermissionRunWrite, log)

	// Assert
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermissionForUpdate_WithMock_NoIdentityInContext(t *testing.T) {
	// Set up mock client (won't be called because identity check happens first)
	mockService := &mockKesselInventoryService{
		checkForUpdateFunc: func(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
			// This should not be called
			t.Fatal("CheckForUpdate should not be called when identity is missing")
			return nil, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context without identity
	ctx := context.Background()
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	// Should fail with "no identity found in context"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no identity found in context")
	assert.False(t, allowed)
}

func TestCheckPermissionForUpdate_WithMock_InvalidIdentityType(t *testing.T) {
	// Set up mock client (won't be called because identity validation happens first)
	mockService := &mockKesselInventoryService{
		checkForUpdateFunc: func(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
			// This should not be called
			t.Fatal("CheckForUpdate should not be called when identity type is invalid")
			return nil, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context with wrong identity type (string instead of XRHID struct)
	ctx := context.WithValue(context.Background(), identity.Key, "not-an-xrhid-struct")
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	// Should fail with "identity in context is not of type XRHID"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity in context is not of type XRHID")
	assert.False(t, allowed)
}

func TestCheckPermissionForUpdate_WithMock_UnsupportedIdentityType(t *testing.T) {
	// Set up mock client (won't be called because user extraction fails first)
	mockService := &mockKesselInventoryService{
		checkForUpdateFunc: func(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
			// This should not be called
			t.Fatal("CheckForUpdate should not be called when identity type is unsupported")
			return nil, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context with unsupported identity type (ServiceAccount instead of User)
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "ServiceAccount",
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	// Should fail with "unsupported identity type"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported identity type")
	assert.Contains(t, err.Error(), "ServiceAccount")
	assert.False(t, allowed)
}

func TestCheckPermissionForUpdate_WithMock_Denied(t *testing.T) {
	// Set up mock to deny permission
	mockService := &mockKesselInventoryService{
		checkForUpdateFunc: func(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
			return &kesselv2.CheckForUpdateResponse{
				Allowed: kesselv2.Allowed_ALLOWED_FALSE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	// Assert
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermissionForUpdate_WithMock_KesselError(t *testing.T) {
	// Set up mock to return error
	mockService := &mockKesselInventoryService{
		checkForUpdateFunc: func(ctx context.Context, req *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
			return nil, errors.New("kessel update check failed")
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Kessel check for update failed")
	assert.Contains(t, err.Error(), "kessel update check failed")
	assert.False(t, allowed)
}

func TestCheckApplicationPermissions_WithMock_AllAllowed(t *testing.T) {
	// Set up mock to allow all permissions
	callCount := 0
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			callCount++
			return &kesselv2.CheckResponse{
				Allowed: kesselv2.Allowed_ALLOWED_TRUE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowedApps, err := CheckApplicationPermissions(ctx, "workspace-123", log)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 3, len(allowedApps)) // All 3 apps allowed
	assert.Contains(t, allowedApps, "config_manager")
	assert.Contains(t, allowedApps, "remediations")
	assert.Contains(t, allowedApps, "tasks")
	assert.Equal(t, 3, callCount) // Should have called Check 3 times
}

func TestCheckApplicationPermissions_WithMock_PartialAllowed(t *testing.T) {
	// Set up mock to allow only some permissions
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			// Only allow remediations
			if req.Relation == PermissionRemediationsRunView {
				return &kesselv2.CheckResponse{
					Allowed: kesselv2.Allowed_ALLOWED_TRUE,
				}, nil
			}
			return &kesselv2.CheckResponse{
				Allowed: kesselv2.Allowed_ALLOWED_FALSE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowedApps, err := CheckApplicationPermissions(ctx, "workspace-123", log)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(allowedApps))
	assert.Contains(t, allowedApps, "remediations")
	assert.NotContains(t, allowedApps, "config_manager")
	assert.NotContains(t, allowedApps, "tasks")
}

func TestCheckApplicationPermissions_WithMock_NoneAllowed(t *testing.T) {
	// Set up mock to deny all permissions
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			return &kesselv2.CheckResponse{
				Allowed: kesselv2.Allowed_ALLOWED_FALSE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowedApps, err := CheckApplicationPermissions(ctx, "workspace-123", log)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, allowedApps)
}

func TestCheckApplicationPermissions_WithMock_ErrorOnFirstCheck(t *testing.T) {
	// Set up mock to return error on first check
	callCount := 0
	mockService := &mockKesselInventoryService{
		checkFunc: func(ctx context.Context, req *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
			callCount++
			// Fail the first check
			if callCount == 1 {
				return nil, errors.New("Kessel service unavailable")
			}
			return &kesselv2.CheckResponse{
				Allowed: kesselv2.Allowed_ALLOWED_TRUE,
			}, nil
		},
	}

	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  identity.User{UserID: "test-user"},
			OrgID: "org-123",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	// Call function
	allowedApps, err := CheckApplicationPermissions(ctx, "workspace-123", log)

	// Should return error immediately on first structural failure
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "structural failure")
	assert.Nil(t, allowedApps)
	assert.Equal(t, 1, callCount) // Should have stopped after first error
}

func TestCheckApplicationPermissions_WithMock_NoIdentity(t *testing.T) {
	// Set up mock client
	mockService := &mockKesselInventoryService{}
	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context with no identity
	ctx := context.Background()
	log := zap.NewNop().Sugar()

	// Call function
	allowedApps, err := CheckApplicationPermissions(ctx, "workspace-123", log)

	// Should return error for missing identity (structural failure)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no identity found in context")
	assert.Nil(t, allowedApps)
}

func TestCheckApplicationPermissions_WithMock_InvalidIdentityType(t *testing.T) {
	// Set up mock client
	mockService := &mockKesselInventoryService{}
	client = newMockClient(mockService)
	defer func() { client = nil }()

	// Context with wrong identity type
	ctx := context.WithValue(context.Background(), identity.Key, "not-an-xrhid")
	log := zap.NewNop().Sugar()

	// Call function
	allowedApps, err := CheckApplicationPermissions(ctx, "workspace-123", log)

	// Should return error for invalid identity type (structural failure)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity in context is not of type XRHID")
	assert.Nil(t, allowedApps)
}
