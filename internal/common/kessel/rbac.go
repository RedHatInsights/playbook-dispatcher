// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/project-kessel/inventory-client-go/common"
	"github.com/spf13/viper"
)

// RbacClient provides access to RBAC service operations
type RbacClient interface {
	GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error)
}

type rbacClientImpl struct {
	baseURL     string
	client      http.Client
	tokenClient *common.TokenClient
}

// NewRbacClient creates a new RBAC client for workspace lookups
func NewRbacClient(cfg *viper.Viper, tokenClient *common.TokenClient) RbacClient {
	timeout := time.Duration(cfg.GetInt("rbac.timeout")) * time.Second

	return &rbacClientImpl{
		baseURL: fmt.Sprintf("%s://%s:%d",
			cfg.GetString("rbac.scheme"),
			cfg.GetString("rbac.host"),
			cfg.GetInt("rbac.port")),
		client: http.Client{
			Timeout: timeout,
		},
		tokenClient: tokenClient,
	}
}

var _ RbacClient = &rbacClientImpl{}

type workspace struct {
	ID string `json:"id"`
}

type workspacesResponse struct {
	Data []workspace `json:"data"`
}

// GetDefaultWorkspaceID retrieves the default workspace ID for an organization from RBAC
// This calls the RBAC v2 API: GET /api/rbac/v2/workspaces/?type=default
func (r *rbacClientImpl) GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error) {
	url := fmt.Sprintf("%s/api/rbac/v2/workspaces/?type=default", r.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("x-rh-rbac-org-id", orgID)

	// Add authentication token if available
	if r.tokenClient != nil {
		token, err := r.tokenClient.GetToken()
		if err != nil {
			return "", fmt.Errorf("error obtaining authentication token: %w", err)
		}

		req.Header.Add("authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read body for both success and error cases
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Handle non-200 responses with detailed diagnostics
	if resp.StatusCode != http.StatusOK {
		// Truncate body snippet for error message (max 200 chars)
		bodySnippet := string(body)
		if len(bodySnippet) > 200 {
			bodySnippet = bodySnippet[:200] + "..."
		}
		return "", fmt.Errorf("RBAC API error: status %d %s, body: %s",
			resp.StatusCode, http.StatusText(resp.StatusCode), bodySnippet)
	}

	var response workspacesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	// Distinguish between "no workspace" and "multiple workspaces"
	if len(response.Data) == 0 {
		return "", fmt.Errorf("no default workspace found for organization %s - workspace may not be configured", orgID)
	}

	if len(response.Data) > 1 {
		return "", fmt.Errorf("multiple default workspaces found for organization %s: got %d, expected 1 - this indicates a data inconsistency", orgID, len(response.Data))
	}

	return response.Data[0].ID, nil
}
