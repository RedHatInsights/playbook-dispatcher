// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"fmt"
	"playbook-dispatcher/internal/common/config"

	"github.com/project-kessel/inventory-client-go/common"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	client      *v1beta2.InventoryClient
	tokenClient *common.TokenClient
	rbacClient  RbacClient
)

// Initialize creates and configures the Kessel inventory client
// This should be called during application startup
func Initialize(cfg *viper.Viper, log *zap.SugaredLogger) error {
	if !cfg.GetBool("kessel.enabled") {
		log.Info("Kessel client disabled")
		return nil
	}

	kesselURL := cfg.GetString("kessel.url")
	if kesselURL == "" {
		return fmt.Errorf("kessel.url is required when kessel.enabled=true")
	}

	log.Infow("Initializing Kessel client",
		"url", kesselURL,
		"auth_enabled", cfg.GetBool("kessel.auth.enabled"),
		"insecure", cfg.GetBool("kessel.insecure"))

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
	client, err = v1beta2.New(kesselConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kessel client: %w", err)
	}

	// Create token client for authentication if enabled
	if cfg.GetBool("kessel.auth.enabled") {
		tokenClient = common.NewTokenClient(kesselConfig)
		log.Info("Kessel authentication enabled")
	}

	// Create RBAC client for workspace lookups
	rbacClient = NewRbacClient(cfg, tokenClient)

	log.Info("Kessel client initialized successfully")
	return nil
}

// GetClient returns the initialized Kessel inventory client
// Returns nil if Kessel is not enabled or not initialized
func GetClient() *v1beta2.InventoryClient {
	return client
}

// GetTokenClient returns the token client for authentication
// Returns nil if authentication is not enabled
func GetTokenClient() *common.TokenClient {
	return tokenClient
}

// GetRbacClient returns the RBAC client for workspace lookups
// Returns nil if Kessel is not enabled or not initialized
func GetRbacClient() RbacClient {
	return rbacClient
}

// IsEnabled returns true if the Kessel client is initialized and ready to use
func IsEnabled() bool {
	return client != nil
}

// Close cleans up the Kessel client resources
// This should be called during application shutdown
func Close() error {
	if client == nil {
		return nil
	}

	// The inventory client doesn't have an explicit Close method
	// but we can clear the references
	client = nil
	tokenClient = nil
	rbacClient = nil

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
