// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"fmt"
	"playbook-dispatcher/internal/common/config"
	"strings"

	"github.com/project-kessel/inventory-client-go/common"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// ClientManager holds all Kessel-related clients (replaces separate global variables)
type ClientManager struct {
	client      *v1beta2.InventoryClient
	tokenClient *common.TokenClient
	rbacClient  RbacClient
}

var globalManager *ClientManager

// Initialize creates and configures the Kessel inventory client
// This should be called during application startup
func Initialize(cfg *viper.Viper, log *zap.SugaredLogger) error {
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

	options := []func(*common.Config){
		common.WithgRPCUrl(kesselURL),
		common.WithTLSInsecure(cfg.GetBool("kessel.insecure")),
	}

	// Add authentication if enabled
	if cfg.GetBool("kessel.auth.enabled") {
		clientID := cfg.GetString("kessel.auth.client.id")
		clientSecret := cfg.GetString("kessel.auth.client.secret")
		oidcIssuer := cfg.GetString("kessel.auth.oidc.issuer")

		if clientID == "" || clientSecret == "" || oidcIssuer == "" {
			return fmt.Errorf("kessel authentication requires client.id, client.secret, and oidc.issuer")
		}

		options = append(options, common.WithAuthEnabled(clientID, clientSecret, oidcIssuer))
	}

	kesselConfig := common.NewConfig(options...)

	var err error
	client, err := v1beta2.New(kesselConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kessel client: %w", err)
	}

	// Create token client for authentication if enabled
	var tokenClient *common.TokenClient
	if cfg.GetBool("kessel.auth.enabled") {
		tokenClient = common.NewTokenClient(kesselConfig)
		log.Info("Kessel authentication enabled")
	}

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

	log.Infow("Creating RBAC client for workspace lookups",
		"rbac_url", rbacURL,
		"rbac_scheme", scheme,
		"rbac_host", host,
		"rbac_port", port)

	rbacClient := NewRbacClient(rbacURL, tokenClient)

	// Store all clients in manager
	globalManager = &ClientManager{
		client:      client,
		tokenClient: tokenClient,
		rbacClient:  rbacClient,
	}

	log.Info("Kessel client initialized successfully")
	return nil
}

// GetClient returns the initialized Kessel inventory client
// Returns nil if Kessel is not enabled or not initialized
func GetClient() *v1beta2.InventoryClient {
	if globalManager == nil {
		return nil
	}
	return globalManager.client
}

// GetTokenClient returns the token client for authentication
// Returns nil if authentication is not enabled
func GetTokenClient() *common.TokenClient {
	if globalManager == nil {
		return nil
	}
	return globalManager.tokenClient
}

// GetRbacClient returns the RBAC client for workspace lookups
// Returns nil if Kessel is not enabled or not initialized
func GetRbacClient() RbacClient {
	if globalManager == nil {
		return nil
	}
	return globalManager.rbacClient
}

// IsEnabled returns true if the Kessel client is initialized and ready to use
func IsEnabled() bool {
	return globalManager != nil && globalManager.client != nil
}

// Close cleans up the Kessel client resources
// This should be called during application shutdown
func Close() error {
	if globalManager == nil {
		return nil
	}

	// The inventory client doesn't have an explicit Close method
	// but we can clear the references
	globalManager = nil

	return nil
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
func SetClientForTesting(client *v1beta2.InventoryClient, tokenClient *common.TokenClient, rbacClient RbacClient) func() {
	oldManager := globalManager
	globalManager = &ClientManager{
		client:      client,
		tokenClient: tokenClient,
		rbacClient:  rbacClient,
	}
	return func() {
		globalManager = oldManager
	}
}
