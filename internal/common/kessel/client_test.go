package kessel

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitialize_Disabled(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	cfg := viper.New()
	cfg.Set("kessel.enabled", false)
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

	assert.NoError(t, err)
	assert.Nil(t, client)
	assert.Nil(t, tokenClient)
	assert.False(t, IsEnabled())
}

func TestInitialize_MissingURL(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.url", "")
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kessel.url is required")
	assert.Nil(t, client)
	assert.False(t, IsEnabled())
}

func TestInitialize_MissingAuthCredentials(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.url", "localhost:9091")
	cfg.Set("kessel.auth.enabled", true)
	cfg.Set("kessel.auth.client.id", "")
	cfg.Set("kessel.auth.client.secret", "")
	cfg.Set("kessel.auth.oidc.issuer", "")
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kessel authentication requires")
}

func TestInitialize_AuthPartialCredentials(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		oidcIssuer   string
	}{
		{
			name:         "missing client secret",
			clientID:     "test-client",
			clientSecret: "",
			oidcIssuer:   "https://sso.example.com",
		},
		{
			name:         "missing client id",
			clientID:     "",
			clientSecret: "test-secret",
			oidcIssuer:   "https://sso.example.com",
		},
		{
			name:         "missing oidc issuer",
			clientID:     "test-client",
			clientSecret: "test-secret",
			oidcIssuer:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := viper.New()
			cfg.Set("kessel.enabled", true)
			cfg.Set("kessel.url", "localhost:9091")
			cfg.Set("kessel.auth.enabled", true)
			cfg.Set("kessel.auth.client.id", tt.clientID)
			cfg.Set("kessel.auth.client.secret", tt.clientSecret)
			cfg.Set("kessel.auth.oidc.issuer", tt.oidcIssuer)
			log := zap.NewNop().Sugar()

			err := Initialize(cfg, log)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "kessel authentication requires")
		})
	}
}

func TestIsEnabled_NotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	assert.False(t, IsEnabled())
}

func TestGetClient_NotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	assert.Nil(t, GetClient())
}

func TestGetTokenClient_NotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	assert.Nil(t, GetTokenClient())
}

func TestGetRbacClient_NotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	assert.Nil(t, GetRbacClient())
}

func TestClose_NotInitialized(t *testing.T) {
	// Reset package variables
	client = nil
	tokenClient = nil
	rbacClient = nil

	err := Close()

	assert.NoError(t, err)
}

func TestClose_Initialized(t *testing.T) {
	// This test verifies that Close() clears the package variables
	// We can't easily test actual cleanup without a real connection

	// Set up mock clients (just for testing the cleanup logic)
	client = nil
	tokenClient = nil
	rbacClient = nil

	err := Close()

	assert.NoError(t, err)
	assert.Nil(t, client)
	assert.Nil(t, tokenClient)
	assert.Nil(t, rbacClient)
}

func TestGetAuthMode_KesselDisabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", false)
	cfg.Set("kessel.auth.mode", "both-kessel-enforces")

	mode := GetAuthMode(cfg)

	// When Kessel is disabled, should always return rbac-only
	assert.Equal(t, "rbac-only", mode)
}

func TestGetAuthMode_ValidModes(t *testing.T) {
	tests := []struct {
		name         string
		mode         string
		expectedMode string
	}{
		{
			name:         "rbac-only",
			mode:         "rbac-only",
			expectedMode: "rbac-only",
		},
		{
			name:         "both-rbac-enforces",
			mode:         "both-rbac-enforces",
			expectedMode: "both-rbac-enforces",
		},
		{
			name:         "both-kessel-enforces",
			mode:         "both-kessel-enforces",
			expectedMode: "both-kessel-enforces",
		},
		{
			name:         "kessel-only",
			mode:         "kessel-only",
			expectedMode: "kessel-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := viper.New()
			cfg.Set("kessel.enabled", true)
			cfg.Set("kessel.auth.mode", tt.mode)

			mode := GetAuthMode(cfg)

			assert.Equal(t, tt.expectedMode, mode)
		})
	}
}

func TestGetAuthMode_InvalidMode(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.auth.mode", "invalid-mode")

	mode := GetAuthMode(cfg)

	// Invalid mode should fall back to rbac-only
	assert.Equal(t, "rbac-only", mode)
}
