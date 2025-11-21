package features

import (
	"context"
	"playbook-dispatcher/internal/common/config"
	"testing"

	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetKesselAuthMode_KesselDisabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", false)
	cfg.Set("unleash.enabled", true)
	log := zap.NewNop().Sugar()

	mode := GetKesselAuthMode(cfg, log)

	assert.Equal(t, config.KesselModeRBACOnly, mode)
}

func TestGetKesselAuthMode_UnleashDisabled_UseEnvVar(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.auth.mode", config.KesselModeBothRBACEnforces)
	cfg.Set("unleash.enabled", false)
	log := zap.NewNop().Sugar()

	mode := GetKesselAuthMode(cfg, log)

	assert.Equal(t, config.KesselModeBothRBACEnforces, mode)
}

func TestGetKesselAuthMode_InvalidMode_ReturnDefault(t *testing.T) {
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.auth.mode", "invalid-mode")
	cfg.Set("unleash.enabled", false)
	log := zap.NewNop().Sugar()

	mode := GetKesselAuthMode(cfg, log)

	assert.Equal(t, config.KesselModeRBACOnly, mode)
}

func TestGetKesselAuthModeWithContext_KesselDisabled(t *testing.T) {
	ctx := context.Background()
	cfg := viper.New()
	cfg.Set("kessel.enabled", false)
	cfg.Set("unleash.enabled", true)
	log := zap.NewNop().Sugar()

	mode := GetKesselAuthModeWithContext(ctx, cfg, log)

	assert.Equal(t, config.KesselModeRBACOnly, mode)
}

func TestGetKesselAuthModeWithContext_UnleashDisabled_UseEnvVar(t *testing.T) {
	ctx := context.Background()
	cfg := viper.New()
	cfg.Set("kessel.enabled", true)
	cfg.Set("kessel.auth.mode", config.KesselModeBothKesselEnforces)
	cfg.Set("unleash.enabled", false)
	log := zap.NewNop().Sugar()

	mode := GetKesselAuthModeWithContext(ctx, cfg, log)

	assert.Equal(t, config.KesselModeBothKesselEnforces, mode)
}

func TestMapVariantToMode_AllVariants(t *testing.T) {
	log := zap.NewNop().Sugar()

	tests := []struct {
		variant  string
		expected string
	}{
		{VariantRBACOnly, config.KesselModeRBACOnly},
		{VariantBothRBACEnforces, config.KesselModeBothRBACEnforces},
		{VariantBothKesselEnforces, config.KesselModeBothKesselEnforces},
		{VariantKesselOnly, config.KesselModeKesselOnly},
		{"disabled", ""},
		{"unknown-variant", ""},
	}

	for _, tt := range tests {
		t.Run(tt.variant, func(t *testing.T) {
			mode := mapVariantToMode(tt.variant, log)
			assert.Equal(t, tt.expected, mode)
		})
	}
}

func TestIsValidMode(t *testing.T) {
	tests := []struct {
		mode     string
		expected bool
	}{
		{config.KesselModeRBACOnly, true},
		{config.KesselModeBothRBACEnforces, true},
		{config.KesselModeBothKesselEnforces, true},
		{config.KesselModeKesselOnly, true},
		{"invalid-mode", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			valid := isValidMode(tt.mode)
			assert.Equal(t, tt.expected, valid)
		})
	}
}

func TestBuildUnleashContext_NoIdentity(t *testing.T) {
	ctx := context.Background()
	log := zap.NewNop().Sugar()

	unleashCtx := buildUnleashContext(ctx, log)

	assert.Equal(t, "", unleashCtx.UserId)
	assert.Empty(t, unleashCtx.Properties)
}

func TestBuildUnleashContext_WithIdentity(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			OrgID: "test-org-123",
			Type:  "User",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	unleashCtx := buildUnleashContext(ctx, log)

	assert.Equal(t, "test-org-123", unleashCtx.UserId)
	assert.Equal(t, "test-org-123", unleashCtx.Properties["orgId"])
}

func TestBuildUnleashContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), identity.Key, "not-an-xrhid")
	log := zap.NewNop().Sugar()

	unleashCtx := buildUnleashContext(ctx, log)

	assert.Equal(t, "", unleashCtx.UserId)
	assert.Empty(t, unleashCtx.Properties)
}

func TestBuildUnleashContext_EmptyOrgID(t *testing.T) {
	xrhid := identity.XRHID{
		Identity: identity.Identity{
			OrgID: "",
			Type:  "User",
		},
	}
	ctx := context.WithValue(context.Background(), identity.Key, xrhid)
	log := zap.NewNop().Sugar()

	unleashCtx := buildUnleashContext(ctx, log)

	assert.Equal(t, "", unleashCtx.UserId)
	assert.Empty(t, unleashCtx.Properties)
}

func TestGetVariantWithFallback_NilVariant(t *testing.T) {
	// This test assumes Unleash is not initialized, so GetVariant returns nil or disabled
	fallbackVariant := "rbac-only"

	variant := GetVariantWithFallback(fallbackVariant)

	assert.NotNil(t, variant)
	assert.Equal(t, fallbackVariant, variant.Name)
	assert.True(t, variant.Enabled)
}
