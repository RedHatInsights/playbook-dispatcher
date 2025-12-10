package kessel

import (
	"context"
	"testing"

	"github.com/project-kessel/inventory-client-go/common"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitialize_Disabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", false)
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

	assert.NoError(t, err)
	assert.Nil(t, globalManager)
	assert.False(t, IsEnabled())
}

func TestInitialize_MissingURL(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.url", "")
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

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

	err := Initialize(cfg, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client.id")
}

func TestGetClient_NotInitialized(t *testing.T) {
	globalManager = nil

	client := GetClient()

	assert.Nil(t, client)
}

func TestGetTokenClient_NotInitialized(t *testing.T) {
	globalManager = nil

	tokenClient := GetTokenClient()

	assert.Nil(t, tokenClient)
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

	// Create mock client
	mockClient := &v1beta2.InventoryClient{}
	mockTokenClient := &common.TokenClient{}
	mockRbacClient := &mockRbacClient{}

	// Use test helper
	cleanup := SetClientForTesting(mockClient, mockTokenClient, mockRbacClient)

	// Verify clients are set
	assert.Equal(t, mockClient, GetClient())
	assert.Equal(t, mockTokenClient, GetTokenClient())
	assert.Equal(t, mockRbacClient, GetRbacClient())
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
		client:      &v1beta2.InventoryClient{},
		tokenClient: &common.TokenClient{},
		rbacClient:  &mockRbacClient{},
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

	err := Initialize(cfg, log)
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

	err := Initialize(cfg, log)
	assert.NoError(t, err)

	// Verify the rbacClient was created (URL should be http://localhost:8080, not http://localhost:8080:9999)
	assert.NotNil(t, globalManager)
	assert.NotNil(t, globalManager.rbacClient)

	err = Close()
	assert.NoError(t, err)
}
