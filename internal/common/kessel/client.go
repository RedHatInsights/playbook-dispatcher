// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"context"
	"fmt"
	"playbook-dispatcher/internal/common/config"
	"strings"
	"time"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	kesselv2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClientManager holds all Kessel-related clients (replaces separate global variables)
type ClientManager struct {
	client     kesselv2.KesselInventoryServiceClient
	conn       *grpc.ClientConn
	tokenCreds *auth.OAuth2ClientCredentials
	rbacClient RbacClient
	kesselClientWithCache *KesselClientWithCache
}

var globalManager *ClientManager

// Initialize creates and configures the Kessel inventory client
// This should be called during application startup
func Initialize(ctx context.Context, cfg *viper.Viper, log *zap.SugaredLogger) error {
	kesselEnabled := cfg.GetBool("kessel.enabled")
	if !kesselEnabled {
		log.Infow("Kessel client disabled",
			"kessel_enabled", kesselEnabled)
		return nil
	}

	kesselURL := cfg.GetString("kessel.url")
	if kesselURL == "" {
		return fmt.Errorf("kessel.url is required when kessel.enabled=true")
	}

	log.Infow("Initializing Kessel client",
		"kessel_enabled", kesselEnabled,
		"kessel_url", kesselURL,
		"kessel_auth_enabled", cfg.GetBool("kessel.auth.enabled"),
		"kessel_insecure", cfg.GetBool("kessel.insecure"),
		"kessel_auth_mode", cfg.GetString("kessel.auth.mode"),
		"kessel_principal_domain", cfg.GetString("kessel.principal.domain"),
		"kessel_auth_oidc_issuer", cfg.GetString("kessel.auth.oidc.issuer"))

	builder := kesselv2.NewClientBuilder(kesselURL)

	var tokenCreds *auth.OAuth2ClientCredentials

	// Configure authentication
	if cfg.GetBool("kessel.auth.enabled") {
		clientID := cfg.GetString("kessel.auth.client.id")
		clientSecret := cfg.GetString("kessel.auth.client.secret")
		oidcIssuer := cfg.GetString("kessel.auth.oidc.issuer")

		if clientID == "" || clientSecret == "" || oidcIssuer == "" {
			return fmt.Errorf("kessel authentication requires client.id, client.secret, and oidc.issuer")
		}

		// Resolve the OIDC token endpoint.
		// The config value (kessel.auth.oidc.issuer) may be either an issuer URL
		// or a direct token endpoint URL (legacy Keycloak default). Handle both.
		tokenEndpoint, err := resolveTokenEndpoint(ctx, oidcIssuer, log)
		if err != nil {
			return fmt.Errorf("failed to resolve OIDC token endpoint: %w", err)
		}

		creds := auth.NewOAuth2ClientCredentials(clientID, clientSecret, tokenEndpoint)
		tokenCreds = &creds

		// Apply TLS configuration: kessel.insecure must also work with auth enabled.
		// Passing insecure.NewCredentials() as channelCredentials ensures the
		// authenticated builder path respects the insecure setting.
		if cfg.GetBool("kessel.insecure") {
			builder.OAuth2ClientAuthenticated(tokenCreds, insecure.NewCredentials())
		} else {
			builder.OAuth2ClientAuthenticated(tokenCreds, nil)
		}
		log.Info("Kessel authentication enabled")
	} else if cfg.GetBool("kessel.insecure") {
		builder.Insecure()
	} else {
		builder.Unauthenticated(nil)
	}

	client, conn, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to create Kessel client: %w", err)
	}

	// Create Kessel client with caching
	kesselClientWithCache := NewKesselClientWithCache(client)

	// Create RBAC client for workspace lookups
	// Build RBAC URL properly, handling cases where host might already contain a port
	host := cfg.GetString("rbac.host")
	port := cfg.GetInt("rbac.port")
	scheme := cfg.GetString("rbac.scheme")

	var rbacURL string
	if strings.Contains(host, ":") {
		// Host already contains a port, don't append another one
		rbacURL = fmt.Sprintf("%s://%s", scheme, host)
	} else {
		// Host doesn't contain a port, append it
		rbacURL = fmt.Sprintf("%s://%s:%d", scheme, host, port)
	}

	rbacTimeout := time.Duration(cfg.GetInt64("rbac.timeout")) * time.Second

	// Build RBAC client configuration
	rbacClientConfig := RbacClientConfig{
		TokenTimeout:       time.Duration(cfg.GetInt64("kessel.token.timeout")) * time.Second,
		TokenMaxRetries:    cfg.GetInt("kessel.token.max_retries"),
		TokenMaxRetriesSet: cfg.IsSet("kessel.token.max_retries"),
	}

	// Avoid nil pointer wrapped in interface gotcha:
	// When kessel.auth.enabled=false, tokenCreds is nil.
	// Passing it directly to NewRbacClient creates a non-nil TokenClient interface
	// (type descriptor exists but value is nil), which passes != nil checks but panics
	// when methods are called. Instead, pass an explicit nil interface.
	var tokenClientInterface TokenClient
	if tokenCreds != nil {
		tokenClientInterface = tokenCreds
	}
	rbacClient := NewRbacClient(rbacURL, tokenClientInterface, rbacTimeout, rbacClientConfig, log)

	// Store all clients in manager
	globalManager = &ClientManager{
		client:                client,
		conn:                  conn,
		tokenCreds:            tokenCreds,
		rbacClient:            rbacClient,
		kesselClientWithCache: kesselClientWithCache,
	}

	log.Info("Kessel client initialized successfully")
	return nil
}

// GetClient returns the initialized Kessel inventory service client
// Returns nil if Kessel is not enabled or not initialized
func GetClient() kesselv2.KesselInventoryServiceClient {
	if globalManager == nil {
		return nil
	}
	return globalManager.client
}

// GetTokenCreds returns the OAuth2 credentials for authentication
// Returns nil if authentication is not enabled
func GetTokenCreds() *auth.OAuth2ClientCredentials {
	if globalManager == nil {
		return nil
	}
	return globalManager.tokenCreds
}

// GetRbacClient returns the RBAC client for workspace lookups
// Returns nil if Kessel is not enabled or not initialized
func GetRbacClient() RbacClient {
	if globalManager == nil {
		return nil
	}
	return globalManager.rbacClient
}

// GetKesselClientWithCache returns the Kessel client with caching
// Returns nil if Kessel is not enabled or not initialized
func GetKesselClientWithCache() *KesselClientWithCache {
	if globalManager == nil {
		return nil
	}
	return globalManager.kesselClientWithCache
}

// IsEnabled returns true if the Kessel client is initialized and ready to use
func IsEnabled() bool {
	return globalManager != nil && globalManager.client != nil
}

const (
	// oidcDiscoveryTimeout is the deadline for each OIDC discovery attempt.
	oidcDiscoveryTimeout = 30 * time.Second
	// oidcDiscoveryRetries is the number of retry attempts for transient discovery failures.
	oidcDiscoveryRetries = 3
	// keycloakTokenSuffix is the standard Keycloak token endpoint path suffix.
	keycloakTokenSuffix = "/protocol/openid-connect/token"
)

// resolveTokenEndpoint determines the OIDC token endpoint from the configured URL.
// If the URL is already a token endpoint (contains the Keycloak token path suffix),
// it is used directly. Otherwise, OIDC discovery is performed with timeout and retry.
func resolveTokenEndpoint(ctx context.Context, configuredURL string, log *zap.SugaredLogger) (string, error) {
	// If the configured URL is already a token endpoint, use it directly.
	// This handles the legacy default where kessel.auth.oidc.issuer is set to
	// the full token endpoint URL rather than the issuer URL.
	if strings.Contains(configuredURL, keycloakTokenSuffix) {
		log.Infow("Using configured URL directly as token endpoint (detected Keycloak token path)",
			"token_endpoint", configuredURL)
		return configuredURL, nil
	}

	// Perform OIDC discovery with timeout and retry for transient failures.
	var lastErr error
	for attempt := 0; attempt <= oidcDiscoveryRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Infow("Retrying OIDC discovery",
				"attempt", attempt+1,
				"backoff", backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", fmt.Errorf("context canceled during OIDC discovery retry: %w", ctx.Err())
			}
		}

		discoveryCtx, cancel := context.WithTimeout(ctx, oidcDiscoveryTimeout)
		discovery, err := auth.FetchOIDCDiscovery(discoveryCtx, configuredURL, auth.FetchOIDCDiscoveryOptions{})
		cancel()

		if err == nil {
			log.Infow("OIDC discovery successful",
				"token_endpoint", discovery.TokenEndpoint)
			return discovery.TokenEndpoint, nil
		}

		lastErr = err
		log.Warnw("OIDC discovery attempt failed",
			"attempt", attempt+1,
			"error", err)
	}

	return "", fmt.Errorf("OIDC discovery failed after %d attempts: %w", oidcDiscoveryRetries+1, lastErr)
}

// Close cleans up the Kessel client resources
// This should be called during application shutdown
func Close() error {
	if globalManager == nil {
		return nil
	}

	var closeErr error
	if globalManager.conn != nil {
		closeErr = globalManager.conn.Close()
	}

	globalManager = nil

	return closeErr
}

// GetAuthMode returns the current Kessel authorization mode from configuration
// This is a convenience wrapper for configuration access
func GetAuthMode(cfg *viper.Viper) string {
	if !cfg.GetBool("kessel.enabled") {
		return config.KesselModeRBACOnly
	}

	mode := cfg.GetString("kessel.auth.mode")
	switch mode {
	case config.KesselModeRBACOnly,
		config.KesselModeBothRBACEnforces,
		config.KesselModeBothKesselEnforces,
		config.KesselModeKesselOnly:
		return mode
	default:
		return config.KesselModeRBACOnly
	}
}

// SetClientForTesting allows tests to inject mock clients
// Returns a cleanup function that restores the original manager
func SetClientForTesting(client kesselv2.KesselInventoryServiceClient, tokenCreds *auth.OAuth2ClientCredentials, rbacClient RbacClient) func() {
	oldManager := globalManager
	globalManager = &ClientManager{
		client:                client,
		tokenCreds:            tokenCreds,
		rbacClient:            rbacClient,
		kesselClientWithCache: NewKesselClientWithCache(client),
	}
	return func() {
		globalManager = oldManager
	}
}
