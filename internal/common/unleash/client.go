package unleash

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Unleash/unleash-go-sdk/v5"
	"github.com/Unleash/unleash-go-sdk/v5/api"
	ucontext "github.com/Unleash/unleash-go-sdk/v5/context"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Initialize initializes the Unleash client
// Returns error if initialization fails, but this is non-fatal - the application can continue
// If Unleash is unavailable, feature flags will fall back to environment variables
func Initialize(cfg *viper.Viper, log *zap.SugaredLogger) error {
	// Check if Unleash is enabled
	if !cfg.GetBool("unleash.enabled") {
		log.Info("Unleash feature flags disabled")
		return nil
	}

	// Validate required configuration
	url := cfg.GetString("unleash.url")
	if url == "" {
		return fmt.Errorf("UNLEASH_URL is required when UNLEASH_ENABLED=true")
	}

	apiToken := cfg.GetString("unleash.api.token")
	if apiToken == "" {
		return fmt.Errorf("UNLEASH_API_TOKEN is required when UNLEASH_ENABLED=true")
	}

	appName := cfg.GetString("unleash.app.name")
	environment := cfg.GetString("unleash.environment")

	log.Infow("Initializing Unleash client",
		"url", url,
		"app_name", appName,
		"environment", environment)

	// Initialize Unleash client
	err := unleash.Initialize(
		// Event listener for logging
		unleash.WithListener(NewDispatcherListener(log)),

		// Application identification
		unleash.WithAppName(appName),
		unleash.WithUrl(url),
		unleash.WithEnvironment(environment),

		// Polling intervals
		unleash.WithRefreshInterval(15*time.Second), // Poll for feature flag updates every 15s
		unleash.WithMetricsInterval(60*time.Second), // Send usage metrics every 60s

		// Authentication
		unleash.WithCustomHeaders(http.Header{
			"Authorization": {apiToken},
		}),
	)

	if err != nil {
		return fmt.Errorf("failed to initialize Unleash client: %w", err)
	}

	log.Info("Unleash client initialized successfully")
	return nil
}

// Close gracefully shuts down the Unleash client
// Should be called during application shutdown (typically with defer)
func Close() {
	unleash.Close()
}

// IsEnabled checks if a feature flag is enabled
// Returns false if Unleash is not initialized
func IsEnabled(featureName string) bool {
	return unleash.IsEnabled(featureName)
}

// IsEnabledWithContext checks if a feature flag is enabled with Unleash context
// Context allows for:
//   - Per-organization gradual rollout
//   - Per-user targeting
//   - Stickiness (same user/org always gets same variant)
func IsEnabledWithContext(featureName string, uctx ucontext.Context) bool {
	return unleash.IsEnabled(featureName, unleash.WithContext(uctx))
}

// GetVariant gets a variant for a feature flag without context
// Returns a variant with Name="disabled" and Enabled=false if:
//   - Unleash is not initialized
//   - Feature flag doesn't exist
//   - Feature flag is disabled
//
// Note: For per-org or per-user targeting, use GetVariantWithContext instead
func GetVariant(featureName string) *api.Variant {
	return unleash.GetVariant(featureName)
}

// GetVariantWithContext gets a variant for a feature flag with Unleash context
// Context allows for:
//   - Per-organization gradual rollout
//   - Per-user targeting
//   - Stickiness (same user/org always gets same variant)
//
// Example context:
//
//	ctx := ucontext.Context{
//	    UserId: orgID,
//	    Properties: map[string]string{"orgId": orgID},
//	}
func GetVariantWithContext(featureName string, uctx ucontext.Context) *api.Variant {
	return unleash.GetVariant(featureName, unleash.WithVariantContext(uctx))
}
