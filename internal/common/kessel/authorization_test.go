package kessel

import (
	"context"
	"errors"
	"testing"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// mockKesselInventoryService is a mock implementation of the Kessel inventory service
type mockKesselInventoryService struct {
	checkResponse          *kesselv2.CheckResponse
	checkError             error
	checkForUpdateResponse *kesselv2.CheckForUpdateResponse
	checkForUpdateError    error
	lastCheckRequest       *kesselv2.CheckRequest
	lastUpdateRequest      *kesselv2.CheckForUpdateRequest
	checkFunc              func(ctx context.Context, in *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error)
}

func (m *mockKesselInventoryService) Check(ctx context.Context, in *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
	m.lastCheckRequest = in
	if m.checkFunc != nil {
		return m.checkFunc(ctx, in, opts...)
	}
	return m.checkResponse, m.checkError
}

func (m *mockKesselInventoryService) CheckForUpdate(ctx context.Context, in *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
	m.lastUpdateRequest = in
	return m.checkForUpdateResponse, m.checkForUpdateError
}

func (m *mockKesselInventoryService) ReportResource(ctx context.Context, in *kesselv2.ReportResourceRequest, opts ...grpc.CallOption) (*kesselv2.ReportResourceResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockKesselInventoryService) DeleteResource(ctx context.Context, in *kesselv2.DeleteResourceRequest, opts ...grpc.CallOption) (*kesselv2.DeleteResourceResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockKesselInventoryService) StreamedListObjects(ctx context.Context, in *kesselv2.StreamedListObjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[kesselv2.StreamedListObjectsResponse], error) {
	return nil, errors.New("not implemented")
}

// Test helper to set up mock client
func setupMockClient(mockService *mockKesselInventoryService) func() {
	mockClient := &v1beta2.InventoryClient{
		KesselInventoryService: mockService,
	}

	cleanup := SetClientForTesting(mockClient, nil, &mockRbacClient{})
	return cleanup
}

func TestCheckPermission_ClientNotInitialized(t *testing.T) {
	globalManager = nil
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(context.Background(), "workspace-123", PermissionRunRead, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "Kessel client not initialized")
}

func TestCheckPermission_Success_Allowed(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkResponse: &kesselv2.CheckResponse{
			Allowed: kesselv2.Allowed_ALLOWED_TRUE,
		},
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	// Create context with identity
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-789", PermissionRunRead, log)

	assert.NoError(t, err)
	assert.True(t, allowed)

	// Verify request details
	assert.NotNil(t, mockService.lastCheckRequest)
	assert.Equal(t, "workspace-789", mockService.lastCheckRequest.Object.ResourceId)
	assert.Equal(t, "workspace", mockService.lastCheckRequest.Object.ResourceType)
	assert.Equal(t, PermissionRunRead, mockService.lastCheckRequest.Relation)
	assert.Equal(t, "redhat/user-123", mockService.lastCheckRequest.Subject.Resource.ResourceId)
	assert.Equal(t, "principal", mockService.lastCheckRequest.Subject.Resource.ResourceType)
}

func TestCheckPermission_Success_Denied(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkResponse: &kesselv2.CheckResponse{
			Allowed: kesselv2.Allowed_ALLOWED_FALSE,
		},
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-789", PermissionRunWrite, log)

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermission_ServiceAccountIdentity(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkResponse: &kesselv2.CheckResponse{
			Allowed: kesselv2.Allowed_ALLOWED_TRUE,
		},
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:           "ServiceAccount",
			ServiceAccount: &identity.ServiceAccount{UserId: "sa-456"},
			OrgID:          "org-789",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, "redhat/sa-456", mockService.lastCheckRequest.Subject.Resource.ResourceId)
}

func TestCheckPermission_KesselError(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkError: errors.New("kessel service unavailable"),
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-789", PermissionRunRead, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "Kessel check failed")
}

func TestCheckPermission_EmptyWorkspaceID(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "", PermissionRunRead, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "workspaceID cannot be empty")
}

func TestCheckPermission_NoIdentityInContext(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	// Context with no identity (empty XRHID returned by GetIdentity)
	ctx := context.Background()
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "failed to extract user ID")
	assert.Contains(t, err.Error(), "unsupported identity type")
}

func TestCheckPermission_UnsupportedIdentityType(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	// Identity with unsupported type
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "System",
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "failed to extract user ID")
	assert.Contains(t, err.Error(), "unsupported identity type: System")
}

func TestCheckPermission_EmptyUserID(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	// User identity with empty UserID
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: ""},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "failed to extract user ID")
	assert.Contains(t, err.Error(), "user ID is empty")
}

func TestCheckPermission_NilUser(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	// User identity with nil User pointer
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  nil,
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermission(ctx, "workspace-123", PermissionRunRead, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "failed to extract user ID")
	assert.Contains(t, err.Error(), "user ID is empty")
}

func TestCheckPermissionForUpdate_Success(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkForUpdateResponse: &kesselv2.CheckForUpdateResponse{
			Allowed: kesselv2.Allowed_ALLOWED_TRUE,
		},
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-789", PermissionRunWrite, log)

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.NotNil(t, mockService.lastUpdateRequest)
	assert.Equal(t, "workspace-789", mockService.lastUpdateRequest.Object.ResourceId)
}

func TestCheckPermissionForUpdate_EmptyWorkspaceID(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "", PermissionRunWrite, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "workspaceID cannot be empty")
}

func TestCheckPermissionForUpdate_KesselError(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkForUpdateError: errors.New("kessel service unavailable"),
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-789", PermissionRunWrite, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "Kessel check for update failed")
}

func TestCheckPermissionForUpdate_NoIdentityInContext(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	// Context with no identity (empty XRHID returned by GetIdentity)
	ctx := context.Background()
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "failed to extract user ID")
	assert.Contains(t, err.Error(), "unsupported identity type")
}

func TestCheckPermissionForUpdate_UnsupportedIdentityType(t *testing.T) {
	mockService := &mockKesselInventoryService{}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	// Identity with unsupported type
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "System",
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	allowed, err := CheckPermissionForUpdate(ctx, "workspace-123", PermissionRunWrite, log)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "failed to extract user ID")
	assert.Contains(t, err.Error(), "unsupported identity type: System")
}

func TestExtractUserID_User(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: &identity.User{UserID: "user-123"},
		},
	}

	userID, err := extractUserID(xrhid)

	assert.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}

func TestExtractUserID_ServiceAccount(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:           "ServiceAccount",
			ServiceAccount: &identity.ServiceAccount{UserId: "sa-456"},
		},
	}

	userID, err := extractUserID(xrhid)

	assert.NoError(t, err)
	assert.Equal(t, "sa-456", userID)
}

func TestExtractUserID_UnsupportedType(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type: "System",
		},
	}

	userID, err := extractUserID(xrhid)

	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "unsupported identity type")
}

func TestExtractUserID_EmptyUserID(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: &identity.User{UserID: ""},
		},
	}

	userID, err := extractUserID(xrhid)

	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "user ID is empty")
}

func TestGetWorkspaceID_Success(t *testing.T) {
	mockRbac := &mockRbacClientWithWorkspace{
		workspaceID: "workspace-123",
	}

	mockClient := &v1beta2.InventoryClient{}
	cleanup := SetClientForTesting(mockClient, nil, mockRbac)
	defer cleanup()

	log := zap.NewNop().Sugar()
	workspaceID, err := GetWorkspaceID(context.Background(), "org-456", log)

	assert.NoError(t, err)
	assert.Equal(t, "workspace-123", workspaceID)
	assert.Equal(t, "org-456", mockRbac.lastOrgID)
}

func TestGetWorkspaceID_RbacError(t *testing.T) {
	mockRbac := &mockRbacClientWithWorkspace{
		err: errors.New("rbac service unavailable"),
	}

	mockClient := &v1beta2.InventoryClient{}
	cleanup := SetClientForTesting(mockClient, nil, mockRbac)
	defer cleanup()

	log := zap.NewNop().Sugar()
	workspaceID, err := GetWorkspaceID(context.Background(), "org-456", log)

	assert.Error(t, err)
	assert.Empty(t, workspaceID)
	assert.Contains(t, err.Error(), "failed to get default workspace ID")
}

func TestGetWorkspaceID_ClientNotInitialized(t *testing.T) {
	globalManager = nil

	log := zap.NewNop().Sugar()
	workspaceID, err := GetWorkspaceID(context.Background(), "org-456", log)

	assert.Error(t, err)
	assert.Empty(t, workspaceID)
	assert.Contains(t, err.Error(), "RBAC client not initialized")
}

func TestCheckPermissions_Success(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkResponse: &kesselv2.CheckResponse{
			Allowed: kesselv2.Allowed_ALLOWED_TRUE,
		},
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	permissions := ServicePermissions{
		Services: []ServicePermission{
			{Name: "config_manager", Permission: "playbook_dispatcher_config_manager_run_view"},
			{Name: "remediations", Permission: "playbook_dispatcher_remediations_run_view"},
			{Name: "tasks", Permission: "playbook_dispatcher_tasks_run_view"},
		},
	}

	allowedApps, err := CheckPermissions(ctx, "workspace-789", permissions, log)

	assert.NoError(t, err)
	assert.Len(t, allowedApps, 3) // All 3 applications
	assert.Contains(t, allowedApps, "config_manager")
	assert.Contains(t, allowedApps, "remediations")
	assert.Contains(t, allowedApps, "tasks")
}

func TestCheckPermissions_PartialAccess(t *testing.T) {
	callCount := 0
	mockService := &mockKesselInventoryService{}

	// Set up response generator that allows only remediations
	mockService.checkFunc = func(ctx context.Context, in *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
		callCount++

		// Only allow remediations
		if in.Relation == "playbook_dispatcher_remediations_run_view" {
			return &kesselv2.CheckResponse{Allowed: kesselv2.Allowed_ALLOWED_TRUE}, nil
		}
		return &kesselv2.CheckResponse{Allowed: kesselv2.Allowed_ALLOWED_FALSE}, nil
	}

	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	permissions := ServicePermissions{
		Services: []ServicePermission{
			{Name: "config_manager", Permission: "playbook_dispatcher_config_manager_run_view"},
			{Name: "remediations", Permission: "playbook_dispatcher_remediations_run_view"},
			{Name: "tasks", Permission: "playbook_dispatcher_tasks_run_view"},
		},
	}

	allowedApps, err := CheckPermissions(ctx, "workspace-789", permissions, log)

	assert.NoError(t, err)
	assert.Len(t, allowedApps, 1)
	assert.Contains(t, allowedApps, "remediations")
	assert.Equal(t, 3, callCount) // Should check all 3 applications
}

func TestCheckPermissions_NoAccess(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkResponse: &kesselv2.CheckResponse{
			Allowed: kesselv2.Allowed_ALLOWED_FALSE,
		},
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	permissions := ServicePermissions{
		Services: []ServicePermission{
			{Name: "config_manager", Permission: "playbook_dispatcher_config_manager_run_view"},
			{Name: "remediations", Permission: "playbook_dispatcher_remediations_run_view"},
			{Name: "tasks", Permission: "playbook_dispatcher_tasks_run_view"},
		},
	}

	allowedApps, err := CheckPermissions(ctx, "workspace-789", permissions, log)

	assert.NoError(t, err)
	assert.Empty(t, allowedApps)
}

func TestCheckPermissions_KesselError(t *testing.T) {
	mockService := &mockKesselInventoryService{
		checkError: errors.New("kessel unavailable"),
	}
	cleanup := setupMockClient(mockService)
	defer cleanup()

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	permissions := ServicePermissions{
		Services: []ServicePermission{
			{Name: "config_manager", Permission: "playbook_dispatcher_config_manager_run_view"},
		},
	}

	allowedApps, err := CheckPermissions(ctx, "workspace-789", permissions, log)

	assert.Error(t, err)
	assert.Nil(t, allowedApps)
	assert.Contains(t, err.Error(), "structural failure")
}

func TestCheckPermissions_ClientNotInitialized(t *testing.T) {
	globalManager = nil

	xrhid := identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			User:  &identity.User{UserID: "user-123"},
			OrgID: "org-456",
		},
	}
	ctx := identity.WithIdentity(context.Background(), xrhid)
	log := zap.NewNop().Sugar()

	permissions := ServicePermissions{
		Services: []ServicePermission{
			{Name: "config_manager", Permission: "playbook_dispatcher_config_manager_run_view"},
		},
	}

	allowedApps, err := CheckPermissions(ctx, "workspace-789", permissions, log)

	assert.Error(t, err)
	assert.Nil(t, allowedApps)
	assert.Contains(t, err.Error(), "cannot perform authorization checks")
}

func TestBuildKesselReferences_Success(t *testing.T) {
	object, subject, err := buildKesselReferences("workspace-123", "redhat/user-456")

	assert.NoError(t, err)
	assert.NotNil(t, object)
	assert.NotNil(t, subject)
	assert.Equal(t, "workspace-123", object.ResourceId)
	assert.Equal(t, "workspace", object.ResourceType)
	assert.Equal(t, "rbac", object.Reporter.Type)
	assert.Equal(t, "redhat/user-456", subject.Resource.ResourceId)
	assert.Equal(t, "principal", subject.Resource.ResourceType)
	assert.Equal(t, "rbac", subject.Resource.Reporter.Type)
}

func TestBuildKesselReferences_EmptyWorkspaceID(t *testing.T) {
	object, subject, err := buildKesselReferences("", "redhat/user-456")

	assert.Error(t, err)
	assert.Nil(t, object)
	assert.Nil(t, subject)
	assert.Contains(t, err.Error(), "workspaceID cannot be empty")
}

func TestBuildKesselReferences_EmptyPrincipalID(t *testing.T) {
	object, subject, err := buildKesselReferences("workspace-123", "")

	assert.Error(t, err)
	assert.Nil(t, object)
	assert.Nil(t, subject)
	assert.Contains(t, err.Error(), "principalID cannot be empty")
}

func TestGetAuthCallOptions_NoTokenClient(t *testing.T) {
	mockClient := &v1beta2.InventoryClient{}
	cleanup := SetClientForTesting(mockClient, nil, nil)
	defer cleanup()

	opts, err := getAuthCallOptions()

	assert.NoError(t, err)
	assert.Empty(t, opts)
}

// mockRbacClientWithWorkspace for testing workspace lookup
type mockRbacClientWithWorkspace struct {
	workspaceID string
	err         error
	lastOrgID   string
}

func (m *mockRbacClientWithWorkspace) GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error) {
	m.lastOrgID = orgID
	return m.workspaceID, m.err
}
