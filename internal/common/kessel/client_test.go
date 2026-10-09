package kessel

import (
	"context"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	kesselv2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestInitialize_Disabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", false)
	log := zap.NewNop().Sugar()

	err := Initialize(context.Background(), cfg, log)

	assert.NoError(t, err)
	assert.Nil(t, globalManager)
	assert.False(t, IsEnabled())
}

func TestInitialize_MissingURL(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.url", "")
	log := zap.NewNop().Sugar()

	err := Initialize(context.Background(), cfg, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kessel.url is required")
}

func TestInitialize_MissingAuthCredentials(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.url", "localhost:9091")
	cfg.Set("kessel.auth.enabled", true)
	cfg.Set("kessel.auth.client.id", "")
	log := zap.NewNop().Sugar()

	err := Initialize(context.Background(), cfg, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client.id")
}

func TestGetClient_NotInitialized(t *testing.T) {
	globalManager = nil

	client := GetClient()

	assert.Nil(t, client)
}

func TestGetTokenCreds_NotInitialized(t *testing.T) {
	globalManager = nil

	tokenCreds := GetTokenCreds()

	assert.Nil(t, tokenCreds)
}

func TestGetRbacClient_NotInitialized(t *testing.T) {
	globalManager = nil

	rbacClient := GetRbacClient()

	assert.Nil(t, rbacClient)
}

func TestIsEnabled_NotInitialized(t *testing.T) {
	globalManager = nil

	enabled := IsEnabled()

	assert.False(t, enabled)
}

func TestSetClientForTesting(t *testing.T) {
	// Save original state
	originalManager := globalManager
	defer func() { globalManager = originalManager }()

	// Create mock client — the mock implements KesselInventoryServiceClient
	mockClient := &mockKesselInventoryServiceMinimal{}
	mockRbac := &mockRbacClient{}

	// Use test helper
	cleanup := SetClientForTesting(mockClient, nil, mockRbac)

	// Verify clients are set
	assert.Equal(t, kesselv2.KesselInventoryServiceClient(mockClient), GetClient())
	assert.Nil(t, GetTokenCreds())
	assert.Equal(t, RbacClient(mockRbac), GetRbacClient())
	assert.True(t, IsEnabled())

	// Call cleanup
	cleanup()

	// Verify original state restored
	assert.Equal(t, originalManager, globalManager)
}

func TestGetAuthMode_KesselDisabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", false)
	cfg.Set("kessel.auth.mode", "kessel-only")

	mode := GetAuthMode(cfg)

	assert.Equal(t, "rbac-only", mode)
}

func TestGetAuthMode_ValidModes(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected string
	}{
		{
			name:     "rbac-only mode",
			mode:     "rbac-only",
			expected: "rbac-only",
		},
		{
			name:     "both-rbac-enforces mode",
			mode:     "both-rbac-enforces",
			expected: "both-rbac-enforces",
		},
		{
			name:     "both-kessel-enforces mode",
			mode:     "both-kessel-enforces",
			expected: "both-kessel-enforces",
		},
		{
			name:     "kessel-only mode",
			mode:     "kessel-only",
			expected: "kessel-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := viper.New()
			cfg.Set("kessel.enabled", true)
			cfg.Set("kessel.auth.mode", tt.mode)

			mode := GetAuthMode(cfg)

			assert.Equal(t, tt.expected, mode)
		})
	}
}

func TestGetAuthMode_InvalidMode_ReturnDefault(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.auth.mode", "invalid-mode")

	mode := GetAuthMode(cfg)

	assert.Equal(t, "rbac-only", mode)
}

func TestClose_NotInitialized(t *testing.T) {
	globalManager = nil

	err := Close()

	assert.NoError(t, err)
}

func TestClose_Initialized(t *testing.T) {
	// Set up a mock manager
	globalManager = &ClientManager{
		client:     &mockKesselInventoryServiceMinimal{},
		tokenCreds: nil,
		rbacClient: &mockRbacClient{},
	}

	err := Close()

	assert.NoError(t, err)
	assert.Nil(t, globalManager)
}

// mockRbacClient for testing
type mockRbacClient struct{}

func (m *mockRbacClient) GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error) {
	return "mock-workspace-id", nil
}

func (m *mockRbacClient) GetDefaultWorkspaceIDWithCache(ctx context.Context, orgID string) (string, error) {
	return "mock-workspace-id", nil
}

// mockKesselInventoryServiceMinimal implements the minimum interface for client_test.go
type mockKesselInventoryServiceMinimal struct{}

func (m *mockKesselInventoryServiceMinimal) Check(ctx context.Context, in *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) CheckSelf(ctx context.Context, in *kesselv2.CheckSelfRequest, opts ...grpc.CallOption) (*kesselv2.CheckSelfResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) CheckForUpdate(ctx context.Context, in *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) CheckForUpdateBulk(ctx context.Context, in *kesselv2.CheckForUpdateBulkRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateBulkResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) CheckBulk(ctx context.Context, in *kesselv2.CheckBulkRequest, opts ...grpc.CallOption) (*kesselv2.CheckBulkResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) CheckSelfBulk(ctx context.Context, in *kesselv2.CheckSelfBulkRequest, opts ...grpc.CallOption) (*kesselv2.CheckSelfBulkResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) ReportResource(ctx context.Context, in *kesselv2.ReportResourceRequest, opts ...grpc.CallOption) (*kesselv2.ReportResourceResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) DeleteResource(ctx context.Context, in *kesselv2.DeleteResourceRequest, opts ...grpc.CallOption) (*kesselv2.DeleteResourceResponse, error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) StreamedListObjects(ctx context.Context, in *kesselv2.StreamedListObjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[kesselv2.StreamedListObjectsResponse], error) {
	return nil, nil
}
func (m *mockKesselInventoryServiceMinimal) StreamedListSubjects(ctx context.Context, in *kesselv2.StreamedListSubjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[kesselv2.StreamedListSubjectsResponse], error) {
	return nil, nil
}

// Verify mockKesselInventoryServiceMinimal implements the interface
var _ kesselv2.KesselInventoryServiceClient = (*mockKesselInventoryServiceMinimal)(nil)
var _ auth.OAuth2ClientCredentials // ensure import is used

func TestInitialize_RbacURLConstruction_HostWithoutPort(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.url", "localhost:9091")
	cfg.Set("kessel.insecure", true)
	cfg.Set("kessel.auth.enabled", false)
	cfg.Set("rbac.scheme", "http")
	cfg.Set("rbac.host", "localhost")
	cfg.Set("rbac.port", 8080)

	log := zap.NewNop().Sugar()

	err := Initialize(context.Background(), cfg, log)
	assert.NoError(t, err)

	// Verify the rbacClient was created with correct URL (scheme://host:port)
	// We can't directly inspect the URL, but we can verify the client was created
	assert.NotNil(t, globalManager)
	assert.NotNil(t, globalManager.rbacClient)

	err = Close()
	assert.NoError(t, err)
}

func TestInitialize_RbacURLConstruction_HostWithPort(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.url", "localhost:9091")
	cfg.Set("kessel.insecure", true)
	cfg.Set("kessel.auth.enabled", false)
	cfg.Set("rbac.scheme", "http")
	cfg.Set("rbac.host", "localhost:8080")
	cfg.Set("rbac.port", 9999) // This should be ignored since host already contains port

	log := zap.NewNop().Sugar()

	err := Initialize(context.Background(), cfg, log)
	assert.NoError(t, err)

	// Verify the rbacClient was created (URL should be http://localhost:8080, not http://localhost:8080:9999)
	assert.NotNil(t, globalManager)
	assert.NotNil(t, globalManager.rbacClient)

	err = Close()
	assert.NoError(t, err)
}
