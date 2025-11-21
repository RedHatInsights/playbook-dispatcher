package unleash

import (
	"testing"

	ucontext "github.com/Unleash/unleash-go-sdk/v5/context"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitialize_Disabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("unleash.enabled", false)
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

	assert.NoError(t, err)
}

func TestInitialize_MissingURL(t *testing.T) {
	cfg := viper.New()
	cfg.Set("unleash.enabled", true)
	cfg.Set("unleash.url", "")
	cfg.Set("unleash.api.token", "test-token")
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNLEASH_URL is required")
}

func TestInitialize_MissingToken(t *testing.T) {
	cfg := viper.New()
	cfg.Set("unleash.enabled", true)
	cfg.Set("unleash.url", "http://test-url")
	cfg.Set("unleash.api.token", "")
	log := zap.NewNop().Sugar()

	err := Initialize(cfg, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNLEASH_API_TOKEN is required")
}

func TestGetVariant_NotInitialized(t *testing.T) {
	// When Unleash is not initialized, GetVariant should return a variant
	// The actual behavior depends on the SDK - it may return nil or a disabled variant
	variant := GetVariant("test-feature")

	// Should not panic
	assert.NotNil(t, variant)
}

func TestGetVariantWithContext_NotInitialized(t *testing.T) {
	ctx := ucontext.Context{
		UserId: "test-org",
	}

	variant := GetVariantWithContext("test-feature", ctx)

	// Should not panic
	assert.NotNil(t, variant)
}

func TestIsEnabled_NotInitialized(t *testing.T) {
	// When Unleash is not initialized, IsEnabled should return false
	enabled := IsEnabled("test-feature")

	// Should return false when not initialized
	assert.False(t, enabled)
}

func TestIsEnabledWithContext_NotInitialized(t *testing.T) {
	ctx := ucontext.Context{
		UserId: "test-org",
	}

	enabled := IsEnabledWithContext("test-feature", ctx)

	// Should return false when not initialized
	assert.False(t, enabled)
}

func TestClose_NotInitialized(t *testing.T) {
	// Should not panic even if not initialized
	assert.NotPanics(t, func() {
		Close()
	})
}
