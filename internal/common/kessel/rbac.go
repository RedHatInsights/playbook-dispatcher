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
	"strings"
	"time"

	"playbook-dispatcher/internal/common/unleash/features"
	"playbook-dispatcher/internal/common/utils"

	"github.com/project-kessel/inventory-client-go/common"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
	"github.com/spf13/viper"
)

// RbacClient provides methods for interacting with the RBAC service
// Used for workspace lookups required by Kessel authorization
type RbacClient interface {
	// GetDefaultWorkspaceID retrieves the default workspace ID for an organization
	GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error)
}

// rbacClientImpl implements RbacClient using the RBAC HTTP API
type rbacClientImpl struct {
	client          *http.Client
	rbacURL         string
	tokenClient     *common.TokenClient
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
func NewRbacClient(rbacURL string, tokenClient *common.TokenClient, timeout time.Duration, cfg *viper.Viper) RbacClient {
	const (
		defaultTokenTimeout    = 3 * time.Second
		minTokenTimeout        = 1 * time.Second
		defaultTokenMaxRetries = 2
		maxTokenMaxRetries     = 5
	)

	// Token timeout configuration (used when feature flag is enabled)
	// Clamp to sane range: [minTokenTimeout, requestTimeout]
	tokenTimeout := time.Duration(cfg.GetInt64("kessel.token.timeout")) * time.Second
	if tokenTimeout <= 0 {
		tokenTimeout = defaultTokenTimeout
	} else if tokenTimeout < minTokenTimeout {
		tokenTimeout = minTokenTimeout
	} else if tokenTimeout > timeout {
		// Can't exceed the overall request timeout
		tokenTimeout = timeout
	}

	// Token retry configuration (used when feature flag is enabled)
	// Clamp to sane range: [0, maxTokenMaxRetries]
	tokenMaxRetries := cfg.GetInt("kessel.token.max_retries")
	if tokenMaxRetries < 0 {
		tokenMaxRetries = defaultTokenMaxRetries
	} else if tokenMaxRetries > maxTokenMaxRetries {
		tokenMaxRetries = maxTokenMaxRetries
	}
	// If explicitly set to 0, allow it (no retries)

	return &rbacClientImpl{
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

	body, statusCode, err := r.doRequestWithRetry(ctx, req)
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
func (r *rbacClientImpl) doRequestWithRetry(ctx context.Context, req *http.Request) ([]byte, int, error) {
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

		// Get logger early for token acquisition logging
		log := utils.GetLogFromContextIfAvailable(ctx)

		// Add authentication token if available
		if r.tokenClient != nil {
			var tokenResp *common.TokenResponse
			var tokenErr error

			// Capture request IDs once for consistency
			reqID := request_id.GetReqID(ctx)
			internalReqID := utils.GetInternalRequestID(ctx)

			// Check if token timeout/retry feature is enabled
			if features.IsTokenTimeoutEnabled(ctx) {
				// NEW: Token acquisition with retry and configurable timeout
				// Configuration is read from kessel.token.timeout and kessel.token.max_retries

				// Retry token acquisition with shorter timeout per attempt
				for tokenAttempt := 0; tokenAttempt <= r.tokenMaxRetries; tokenAttempt++ {
					// Check if parent context was canceled
					if ctx.Err() != nil {
						cancel()
						return nil, 0, fmt.Errorf("upstream request canceled, token acquisition aborted: %w", ctx.Err())
					}

					tokenStart := time.Now()

					// Log token acquisition start
					if log != nil {
						log.Debugw("OIDC token acquisition started",
							"request_id", reqID,
							"internal_request_id", internalReqID,
							"token_attempt", tokenAttempt+1,
							"workspace_attempt", attempt+1,
							"token_timeout_enabled", true)
					}

					// Create context with timeout just for token request
					tokenCtx, tokenCancel := context.WithTimeout(ctx, r.tokenTimeout)
					tokenResp, tokenErr = r.tokenClient.GetTokenWithContext(tokenCtx)
					tokenDuration := time.Since(tokenStart)
					tokenCancel() // Clean up immediately

					if tokenErr == nil {
						// Success!
						if log != nil {
							log.Debugw("OIDC token acquisition succeeded",
								"request_id", reqID,
								"internal_request_id", internalReqID,
								"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
								"token_attempt", tokenAttempt+1,
								"workspace_attempt", attempt+1,
								"token_timeout_enabled", true)
						}
						break // Exit retry loop
					}

					// Token acquisition failed - distinguish error types
					// Check if parent context was canceled during token request
					if ctx.Err() != nil {
						if log != nil {
							log.Warnw("OIDC token acquisition failed due to upstream context cancellation",
								"request_id", reqID,
								"internal_request_id", internalReqID,
								"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
								"token_attempt", tokenAttempt+1,
								"workspace_attempt", attempt+1,
								"upstream_error", ctx.Err(),
								"token_error", tokenErr,
								"token_timeout_enabled", true)
						}
						cancel()
						return nil, 0, fmt.Errorf("upstream request canceled during token acquisition: %w", ctx.Err())
					}

					// Token request failed (timeout, network, TLS, etc.) but parent context still valid
					if log != nil {
						log.Warnw("OIDC token acquisition failed",
							"request_id", reqID,
							"internal_request_id", internalReqID,
							"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
							"token_attempt", tokenAttempt+1,
							"workspace_attempt", attempt+1,
							"error", tokenErr,
							"token_timeout_enabled", true)
					}

					// If this was the last retry, give up
					if tokenAttempt == r.tokenMaxRetries {
						cancel()
						// Categorize error for clearer alerting
						errorCategory := categorizeTokenError(tokenErr)
						return nil, 0, fmt.Errorf("token acquisition failed after %d attempts (%s): %w", r.tokenMaxRetries+1, errorCategory, tokenErr)
					}

					// Brief backoff before retry (50-100ms)
					backoff := time.Duration(50+rand.Intn(50)) * time.Millisecond
					time.Sleep(backoff)
				}
			} else {
				// ORIGINAL: Legacy token acquisition (default)
				tokenStart := time.Now()

				// Log token acquisition start with request IDs
				if log != nil {
					log.Debugw("OIDC token acquisition started",
						"request_id", reqID,
						"internal_request_id", internalReqID,
						"attempt", attempt+1,
						"token_timeout_enabled", false)
				}

				tokenResp, tokenErr = r.tokenClient.GetToken()
				tokenDuration := time.Since(tokenStart)

				if tokenErr != nil {
					// Log token acquisition failure
					if log != nil {
						log.Errorw("OIDC token acquisition failed",
							"request_id", reqID,
							"internal_request_id", internalReqID,
							"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
							"attempt", attempt+1,
							"error", tokenErr,
							"token_timeout_enabled", false)
					}
					cancel() // Clean up timeout context before returning error
					return nil, 0, fmt.Errorf("failed to get auth token: %w", tokenErr)
				}

				// Log token acquisition success
				if log != nil {
					log.Debugw("OIDC token acquisition succeeded",
						"request_id", reqID,
						"internal_request_id", internalReqID,
						"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
						"attempt", attempt+1,
						"token_timeout_enabled", false)
				}
			}

			reqWithTimeout.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResp.AccessToken))
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

// errorPatterns maps error substrings to categories for alerting
var errorPatterns = []struct {
	pattern  string
	category string
}{
	{"TLS handshake timeout", "tls-timeout"},
	{"context deadline exceeded", "timeout"},
	{"i/o timeout", "io-timeout"},
	{"connection refused", "connection-refused"},
	{"connection reset", "connection-reset"},
	{"no such host", "dns-error"},
	{"context canceled", "canceled"},
}

// categorizeTokenError inspects the error and returns a category for alerting
func categorizeTokenError(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()

	// Check patterns in order (most specific first)
	for _, p := range errorPatterns {
		if strings.Contains(errStr, p.pattern) {
			return p.category
		}
	}

	return "network-error"
}
