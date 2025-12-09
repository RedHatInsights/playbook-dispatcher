package features

import (
	"context"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/unleash"

	"github.com/Unleash/unleash-go-sdk/v5/api"
	ucontext "github.com/Unleash/unleash-go-sdk/v5/context"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	// KesselFeatureFlag is the name of the Unleash feature flag for Kessel authorization
	KesselFeatureFlag = "playbook-dispatcher-kessel"

	// Variant names matching Unleash dashboard configuration
	VariantRBACOnly           = "rbac-only"
	VariantBothRBACEnforces   = "both-rbac-enforces"
	VariantBothKesselEnforces = "both-kessel-enforces"
	VariantKesselOnly         = "kessel-only"
)

// GetKesselAuthMode determines the current Kessel authorization mode WITHOUT context
// Use GetKesselAuthModeWithContext for per-org or per-user targeting
//
// Mode selection priority:
//  1. If KESSEL_ENABLED=false, always return "rbac-only"
//  2. If UNLEASH_ENABLED=true, get mode from Unleash variant
//  3. If Unleash unavailable or variant disabled, use KESSEL_AUTH_MODE environment variable
//  4. If KESSEL_AUTH_MODE is invalid, return "rbac-only" as safe default
//
// Returns one of: "rbac-only", "both-rbac-enforces", "both-kessel-enforces", "kessel-only"
func GetKesselAuthMode(cfg *viper.Viper, log *zap.SugaredLogger) string {
	// Priority 1: If Kessel not enabled, always use RBAC-only mode
	if !cfg.GetBool("kessel.enabled") {
		return config.KesselModeRBACOnly
	}

	// Priority 2: If Unleash enabled, try to get mode from variant
	if cfg.GetBool("unleash.enabled") {
		variant := unleash.GetVariant(KesselFeatureFlag)

		// Check if variant is enabled and valid
		if variant != nil && variant.Enabled {
			mode := mapVariantToMode(variant.Name, log)
			if mode != "" {
				log.Infow("Using Kessel auth mode from Unleash variant",
					"variant", variant.Name,
					"mode", mode)
				return mode
			}
		} else {
			log.Warnw("Unleash variant not enabled or not found, falling back to KESSEL_AUTH_MODE",
				"feature_flag", KesselFeatureFlag)
		}
	}

	// Priority 3: Use environment variable fallback
	mode := cfg.GetString("kessel.auth.mode")
	log.Infow("Using Kessel auth mode from environment variable",
		"mode", mode)

	// Validate mode
	if !isValidMode(mode) {
		log.Errorw("Invalid KESSEL_AUTH_MODE, falling back to rbac-only",
			"invalid_mode", mode)
		return config.KesselModeRBACOnly
	}

	return mode
}

// GetKesselAuthModeWithContext determines the Kessel authorization mode WITH Unleash context
// This enables:
//   - Per-organization gradual rollout
//   - Per-user targeting
//   - Stickiness (same org/user always gets same variant)
//
// The orgID is extracted from the context and passed to Unleash for targeting decisions.
//
// Example usage:
//
//	mode := features.GetKesselAuthModeWithContext(ctx, cfg, log)
//
// Mode selection priority:
//  1. If KESSEL_ENABLED=false, always return "rbac-only"
//  2. If UNLEASH_ENABLED=true, get mode from Unleash variant with context
//  3. If Unleash unavailable or variant disabled, use KESSEL_AUTH_MODE environment variable
//  4. If KESSEL_AUTH_MODE is invalid, return "rbac-only" as safe default
//
// Returns one of: "rbac-only", "both-rbac-enforces", "both-kessel-enforces", "kessel-only"
func GetKesselAuthModeWithContext(ctx context.Context, cfg *viper.Viper, log *zap.SugaredLogger) string {
	// Priority 1: If Kessel not enabled, always use RBAC-only mode
	if !cfg.GetBool("kessel.enabled") {
		return config.KesselModeRBACOnly
	}

	// Priority 2: If Unleash enabled, try to get mode from variant with context
	if cfg.GetBool("unleash.enabled") {
		// Build Unleash context from request context
		unleashCtx := buildUnleashContext(ctx, log)

		// Get variant with context for targeting
		variant := unleash.GetVariantWithContext(KesselFeatureFlag, unleashCtx)

		// Check if variant is enabled and valid
		if variant != nil && variant.Enabled {
			mode := mapVariantToMode(variant.Name, log)
			if mode != "" {
				log.Infow("Using Kessel auth mode from Unleash variant with context",
					"variant", variant.Name,
					"mode", mode,
					"org_id", unleashCtx.UserId)
				return mode
			}
		} else {
			log.Warnw("Unleash variant not enabled or not found, falling back to KESSEL_AUTH_MODE",
				"feature_flag", KesselFeatureFlag,
				"org_id", unleashCtx.UserId)
		}
	}

	// Priority 3: Use environment variable fallback
	mode := cfg.GetString("kessel.auth.mode")
	log.Infow("Using Kessel auth mode from environment variable",
		"mode", mode)

	// Validate mode
	if !isValidMode(mode) {
		log.Errorw("Invalid KESSEL_AUTH_MODE, falling back to rbac-only",
			"invalid_mode", mode)
		return config.KesselModeRBACOnly
	}

	return mode
}

// buildUnleashContext extracts organization ID from request context and builds Unleash context
// This is used for per-org targeting and gradual rollout
func buildUnleashContext(ctx context.Context, log *zap.SugaredLogger) ucontext.Context {
	// Extract identity from context
	// Note: In v2, GetIdentity returns an empty XRHID if identity is not in context
	xrhid := identity.GetIdentity(ctx)

	// Check if we got a valid identity (non-empty Type indicates presence)
	if xrhid.Identity.Type == "" {
		log.Debug("No identity found in context, using empty Unleash context")
		return ucontext.Context{}
	}

	orgID := xrhid.Identity.OrgID
	if orgID == "" {
		log.Warn("Identity present but OrgID is empty, using empty Unleash context")
		return ucontext.Context{}
	}

	// Build Unleash context with org ID
	return ucontext.Context{
		UserId: orgID, // Use orgID as userId for targeting
		Properties: map[string]string{
			"orgId": orgID,
		},
	}
}

// mapVariantToMode maps Unleash variant names to authorization mode constants
// Returns empty string if variant is unknown (triggers fallback to env var)
func mapVariantToMode(variantName string, log *zap.SugaredLogger) string {
	switch variantName {
	case VariantRBACOnly:
		return config.KesselModeRBACOnly
	case VariantBothRBACEnforces:
		return config.KesselModeBothRBACEnforces
	case VariantBothKesselEnforces:
		return config.KesselModeBothKesselEnforces
	case VariantKesselOnly:
		return config.KesselModeKesselOnly
	case "disabled":
		// Unleash returns variant with name "disabled" when feature is disabled
		return ""
	default:
		log.Warnw("Unknown Unleash variant name, falling back to environment variable",
			"variant", variantName,
			"feature_flag", KesselFeatureFlag)
		return ""
	}
}

// isValidMode checks if a mode string is one of the valid authorization modes
func isValidMode(mode string) bool {
	validModes := map[string]bool{
		config.KesselModeRBACOnly:           true,
		config.KesselModeBothRBACEnforces:   true,
		config.KesselModeBothKesselEnforces: true,
		config.KesselModeKesselOnly:         true,
	}
	return validModes[mode]
}

// GetVariantWithFallback gets the Kessel variant with an explicit fallback
// Useful for testing or when you want fine-grained control over fallback behavior
//
// If the feature flag is disabled or unavailable, returns a mock variant
// with the specified fallback variant name
func GetVariantWithFallback(fallbackVariant string) *api.Variant {
	variant := unleash.GetVariant(KesselFeatureFlag)

	// If variant is valid and enabled, return it
	if variant != nil && variant.Enabled {
		return variant
	}

	// Return a mock variant for fallback
	return &api.Variant{
		Name:    fallbackVariant,
		Enabled: true,
	}
}
