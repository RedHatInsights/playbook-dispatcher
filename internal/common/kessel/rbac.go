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
	"time"

	"playbook-dispatcher/internal/common/unleash/features"
	"playbook-dispatcher/internal/common/utils"

	"github.com/project-kessel/inventory-client-go/common"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
	"go.uber.org/zap"
)

// RbacClientConfig holds configuration for RBAC client
type RbacClientConfig struct {
	// TokenTimeout is the timeout for individual token requests (when feature flag is enabled)
	// Zero or negative values will use the default (3 seconds)
	TokenTimeout time.Duration

	// TokenMaxRetries is the maximum number of retry attempts for token acquisition
	// Negative values will use the default (2 retries)
	// Explicitly set to 0 for no retries (single attempt only)
	// Values > 5 will be clamped to 5
	TokenMaxRetries int

	// TokenMaxRetriesSet indicates if TokenMaxRetries was explicitly set
	// This allows distinguishing between "not set" and "explicitly set to 0"
	TokenMaxRetriesSet bool
}

// Token configuration constants for RBAC client
const (
	// DefaultTokenTimeout is the default timeout for individual token requests
	DefaultTokenTimeout = 3 * time.Second

	// MinTokenTimeout is the minimum allowed timeout for token requests
	MinTokenTimeout = 1 * time.Second

	// DefaultTokenMaxRetries is the default number of retry attempts for token acquisition
	DefaultTokenMaxRetries = 2

	// MaxTokenMaxRetries is the maximum allowed number of retry attempts for token acquisition
	MaxTokenMaxRetries = 5
)

// RbacClient provides methods for interacting with the RBAC service
// Used for workspace lookups required by Kessel authorization
type RbacClient interface {
	// GetDefaultWorkspaceID retrieves the default workspace ID for an organization
	GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error)
}

// TokenClient interface wraps the Kessel token client for testability
type TokenClient interface {
	// GetToken acquires an OIDC token without context (legacy behavior)
	GetToken() (*common.TokenResponse, error)
	// GetTokenWithContext acquires an OIDC token with context support for timeout/cancellation
	GetTokenWithContext(ctx context.Context) (*common.TokenResponse, error)
}

// rbacClientImpl implements RbacClient using the RBAC HTTP API
type rbacClientImpl struct {
	client          *http.Client
	rbacURL         string
	tokenClient     TokenClient
	maxRetries      int
	initialBackoff  time.Duration
	maxBackoff      time.Duration
	requestTimeout  time.Duration
	tokenTimeout    time.Duration // Timeout for individual token requests
	tokenMaxRetries int           // Max retry attempts for token acquisition
}

// rbacWorkspaceResponse represents the RBAC API response for workspace queries
type rbacWorkspaceResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// NewRbacClient creates a new RBAC client for workspace lookups
func NewRbacClient(rbacURL string, tokenClient TokenClient, timeout time.Duration, cfg RbacClientConfig, log *zap.SugaredLogger) RbacClient {
	// Token timeout configuration (used when feature flag is enabled)
	// Values <= 0 use default; otherwise clamp to [MinTokenTimeout, requestTimeout]
	tokenTimeout := cfg.TokenTimeout
	if tokenTimeout <= 0 {
		tokenTimeout = DefaultTokenTimeout
	} else if tokenTimeout < MinTokenTimeout {
		tokenTimeout = MinTokenTimeout
	} else if tokenTimeout > timeout {
		// Can't exceed the overall request timeout
		tokenTimeout = timeout
	}

	// Token retry configuration (used when feature flag is enabled)
	// If not set: use default
	// If set to negative: use default (invalid value)
	// If set to 0: no retries (1 attempt only)
	// If set to 1-5: use that value
	// If set > 5: clamp to 5
	tokenMaxRetries := DefaultTokenMaxRetries
	if cfg.TokenMaxRetriesSet {
		if cfg.TokenMaxRetries < 0 {
			// Negative is invalid, use default
			tokenMaxRetries = DefaultTokenMaxRetries
		} else if cfg.TokenMaxRetries > MaxTokenMaxRetries {
			// Too large, clamp to max
			tokenMaxRetries = MaxTokenMaxRetries
		} else {
			// Valid range [0, MaxTokenMaxRetries], use as-is
			tokenMaxRetries = cfg.TokenMaxRetries
		}
	}

	impl := &rbacClientImpl{
		client: &http.Client{
			Timeout: timeout,
		},
		rbacURL:         rbacURL,
		tokenClient:     tokenClient,
		maxRetries:      3,
		initialBackoff:  100 * time.Millisecond,
		maxBackoff:      2 * time.Second,
		requestTimeout:  timeout,
		tokenTimeout:    tokenTimeout,
		tokenMaxRetries: tokenMaxRetries,
	}

	// Log the clamped configuration values
	if log != nil {
		log.Infow("Created RBAC client with configuration",
			"rbac_url", rbacURL,
			"rbac_timeout_seconds", timeout.Seconds(),
			"token_timeout_seconds", tokenTimeout.Seconds(),
			"token_max_retries", tokenMaxRetries)
	}

	return impl
}

// GetDefaultWorkspaceID retrieves the default workspace ID for an organization
// Returns the workspace ID or an error if the request fails
func (r *rbacClientImpl) GetDefaultWorkspaceID(ctx context.Context, orgID string) (workspaceID string, err error) {
	start := time.Now()

	// Diagnostic: Check if parent context is already canceled before we start
	log := utils.GetLogFromContextIfAvailable(ctx)
	if ctx.Err() != nil {
		if log != nil {
			log.Debugw("Parent context already canceled before workspace lookup",
				"org_id", orgID,
				"error", ctx.Err())
		}
	}

	// Evaluate feature flag once per request to avoid repeated context building on retries
	tokenTimeoutEnabled := features.IsTokenTimeoutEnabled(ctx)

	defer func() {
		if log != nil {
			if err != nil {
				log.Debugw("RBAC workspace lookup failed",
					"duration_ms", time.Since(start).Milliseconds(),
					"org_id", orgID,
					"error", err)
			} else {
				log.Debugw("RBAC workspace lookup succeeded",
					"duration_ms", time.Since(start).Milliseconds(),
					"org_id", orgID,
					"workspace_id", workspaceID)
			}
		}
	}()

	requestURL := fmt.Sprintf("%s/api/rbac/v2/workspaces/?type=default", r.rbacURL)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add org ID header for RBAC v2 authentication
	req.Header.Set("x-rh-rbac-org-id", orgID)

	// Add request ID headers for traceability
	utils.PropagateRequestIDs(ctx, req)

	body, statusCode, err := r.doRequestWithRetry(ctx, req, tokenTimeoutEnabled)
	if err != nil {
		return "", err
	}

	if log != nil {
		log.Debugw("RBAC workspace API response received",
			"status_code", statusCode,
			"response_body", string(body),
			"org_id", orgID)
	}

	var workspaceResp rbacWorkspaceResponse
	if err := json.Unmarshal(body, &workspaceResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(workspaceResp.Data) != 1 {
		return "", fmt.Errorf("unexpected number of default workspaces: %d", len(workspaceResp.Data))
	}

	return workspaceResp.Data[0].ID, nil
}

// doRequestWithRetry executes an HTTP request with retry logic and exponential backoff
// Returns the response body bytes, status code, and any error
// tokenTimeoutEnabled is evaluated once per request to avoid repeated feature flag checks
func (r *rbacClientImpl) doRequestWithRetry(ctx context.Context, req *http.Request, tokenTimeoutEnabled bool) ([]byte, int, error) {
	var lastErr error

	// Diagnostic: Track parent context cancellation
	// Use a done channel to ensure the goroutine exits when the function returns
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			log := utils.GetLogFromContextIfAvailable(ctx)
			if log != nil {
				log.Debugw("Parent context canceled during RBAC workspace lookup",
					"error", ctx.Err())
			}
		case <-done:
			// Function returned, exit goroutine
			return
		}
	}()

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		// Check if parent context was canceled before starting attempt
		if ctx.Err() != nil {
			return nil, 0, fmt.Errorf("request canceled: %w", ctx.Err())
		}

		// Apply per-request timeout (defense-in-depth)
		requestCtx, cancel := context.WithTimeout(ctx, r.requestTimeout)

		reqWithTimeout := req.Clone(requestCtx)

		// Get logger for this attempt
		log := utils.GetLogFromContextIfAvailable(ctx)

		// Acquire authentication token if token client is configured
		var accessToken string
		var err error
		if r.tokenClient != nil {
			// Use pre-evaluated feature flag to avoid repeated context building
			if tokenTimeoutEnabled {
				// New behavior: retry with configurable timeout
				reqID := request_id.GetReqID(ctx)
				internalReqID := utils.GetInternalRequestID(ctx)

				if log != nil {
					log.Debugw("OIDC token acquisition started",
						"request_id", reqID,
						"internal_request_id", internalReqID,
						"token_timeout_enabled", true)
				}

				accessToken, err = r.acquireTokenWithRetry(ctx, reqID, internalReqID, attempt+1, log)
				if err != nil {
					cancel() // Clean up timeout context before returning error
					return nil, 0, err
				}
			} else {
				// Legacy behavior: single attempt, no timeout
				tokenStart := time.Now()
				reqID := request_id.GetReqID(ctx)
				internalReqID := utils.GetInternalRequestID(ctx)

				if log != nil {
					log.Debugw("OIDC token acquisition started",
						"request_id", reqID,
						"internal_request_id", internalReqID,
						"token_timeout_enabled", false)
				}

				tokenResp, tokenErr := r.tokenClient.GetToken()
				tokenDuration := time.Since(tokenStart)

				if tokenErr != nil {
					if log != nil {
						log.Errorw("OIDC token acquisition failed",
							"request_id", reqID,
							"internal_request_id", internalReqID,
							"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
							"error", tokenErr,
							"token_timeout_enabled", false)
					}
					cancel() // Clean up timeout context before returning error
					return nil, 0, fmt.Errorf("OIDC token acquisition failed: %w", tokenErr)
				}

				if log != nil {
					log.Debugw("OIDC token acquisition succeeded",
						"request_id", reqID,
						"internal_request_id", internalReqID,
						"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
						"token_timeout_enabled", false)
				}

				accessToken = tokenResp.AccessToken
			}
		}

		// Add token to request if acquired
		if accessToken != "" {
			reqWithTimeout.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		}

		// Debug logging for outgoing request
		if log != nil {
			headers := make(map[string]string)
			for k, v := range reqWithTimeout.Header {
				if len(v) > 0 {
					// Redact sensitive headers
					if k == "Authorization" {
						headers[k] = "Bearer [REDACTED]"
					} else {
						headers[k] = v[0]
					}
				}
			}
			log.Debugw("Sending RBAC workspace lookup request",
				"method", reqWithTimeout.Method,
				"url", reqWithTimeout.URL.String(),
				"headers", headers,
				"attempt", attempt+1,
				"max_retries", r.maxRetries+1)
		}

		resp, err := r.client.Do(reqWithTimeout)

		// Diagnostic: Check if request context was canceled
		if requestCtx.Err() != nil && log != nil {
			log.Debugw("Request context canceled",
				"attempt", attempt+1,
				"request_error", requestCtx.Err(),
				"parent_error", ctx.Err())
		}

		// Success case
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Read body while context is still alive
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			cancel() // Clean up context immediately after body is consumed

			if readErr != nil {
				return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", readErr)
			}

			return body, resp.StatusCode, nil
		}

		// Store error for later
		var statusCode int
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			statusCode = 0
		} else {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			statusCode = resp.StatusCode
			resp.Body.Close()
		}

		// Check if we should retry
		if !r.shouldRetry(resp, err) || attempt == r.maxRetries {
			cancel() // Clean up timeout context before returning error
			return nil, statusCode, lastErr
		}

		// Clean up this attempt's context before sleeping and retrying
		cancel()

		// Check if parent context was canceled before sleeping
		if ctx.Err() != nil {
			return nil, 0, fmt.Errorf("request canceled during retry: %w", ctx.Err())
		}

		// Calculate backoff with jitter
		backoff := r.calculateBackoff(attempt)
		time.Sleep(backoff)
	}

	return nil, 0, lastErr
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
