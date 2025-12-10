// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/project-kessel/inventory-client-go/common"
)

// RbacClient provides methods for interacting with the RBAC service
// Used for workspace lookups required by Kessel authorization
type RbacClient interface {
	// GetDefaultWorkspaceID retrieves the default workspace ID for an organization
	GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error)
}

// rbacClientImpl implements RbacClient using the RBAC HTTP API
type rbacClientImpl struct {
	client         *http.Client
	rbacURL        string
	tokenClient    *common.TokenClient
	maxRetries     int
	initialBackoff time.Duration
	maxBackoff     time.Duration
	requestTimeout time.Duration
}

// rbacWorkspaceResponse represents the RBAC API response for workspace queries
type rbacWorkspaceResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// NewRbacClient creates a new RBAC client for workspace lookups
func NewRbacClient(rbacURL string, tokenClient *common.TokenClient) RbacClient {
	return &rbacClientImpl{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		rbacURL:        rbacURL,
		tokenClient:    tokenClient,
		maxRetries:     3,
		initialBackoff: 100 * time.Millisecond,
		maxBackoff:     2 * time.Second,
		requestTimeout: 10 * time.Second,
	}
}

// GetDefaultWorkspaceID retrieves the default workspace ID for an organization
// Returns the workspace ID or an error if the request fails
func (r *rbacClientImpl) GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error) {
	requestURL := fmt.Sprintf("%s/api/rbac/v2/workspaces/?org_id=%s&type=default", r.rbacURL, url.QueryEscape(orgID))

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.doRequestWithRetry(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var workspaceResp rbacWorkspaceResponse
	if err := json.Unmarshal(body, &workspaceResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(workspaceResp.Data) == 0 {
		return "", fmt.Errorf("no default workspace found for org %s", orgID)
	}

	return workspaceResp.Data[0].ID, nil
}

// doRequestWithRetry executes an HTTP request with retry logic and exponential backoff
func (r *rbacClientImpl) doRequestWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		// Check if parent context was canceled before starting attempt
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request canceled: %w", ctx.Err())
		}

		// Apply per-request timeout (defense-in-depth)
		requestCtx, cancel := context.WithTimeout(ctx, r.requestTimeout)

		reqWithTimeout := req.Clone(requestCtx)

		// Add authentication token if available
		if r.tokenClient != nil {
			tokenResp, err := r.tokenClient.GetToken()
			if err != nil {
				cancel() // Clean up timeout context before returning error
				return nil, fmt.Errorf("failed to get auth token: %w", err)
			}
			reqWithTimeout.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResp.AccessToken))
		}

		resp, err := r.client.Do(reqWithTimeout)

		// Success case
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			cancel() // Clean up timeout context
			return resp, nil
		}

		// Store error for later
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
		} else {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			resp.Body.Close()
		}

		// Check if we should retry
		if !r.shouldRetry(resp, err) || attempt == r.maxRetries {
			cancel() // Clean up timeout context before returning error
			return nil, lastErr
		}

		// Clean up this attempt's context before sleeping and retrying
		cancel()

		// Check if parent context was canceled before sleeping
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request canceled during retry: %w", ctx.Err())
		}

		// Calculate backoff with jitter
		backoff := r.calculateBackoff(attempt)
		time.Sleep(backoff)
	}

	return nil, lastErr
}

// shouldRetry determines if a request should be retried based on the error or response
func (r *rbacClientImpl) shouldRetry(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on specific HTTP status codes
	if resp != nil {
		// 5xx server errors
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			return true
		}
		// 429 Too Many Requests
		if resp.StatusCode == 429 {
			return true
		}
	}

	// Don't retry on client errors (4xx except 429)
	return false
}

// calculateBackoff calculates the backoff duration with exponential backoff and jitter
// Jitter range: 50-100% of the calculated backoff to prevent thundering herd
func (r *rbacClientImpl) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: initialBackoff * 2^attempt
	backoff := r.initialBackoff * (1 << uint(attempt))

	// Cap at maxBackoff
	if backoff > r.maxBackoff {
		backoff = r.maxBackoff
	}

	// Apply jitter: multiply by random value in range [0.5, 1.0]
	jitter := 0.5 + rand.Float64()*0.5
	return time.Duration(float64(backoff) * jitter)
}
