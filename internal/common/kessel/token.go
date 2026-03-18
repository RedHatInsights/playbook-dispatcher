// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/project-kessel/inventory-client-go/common"
	"go.uber.org/zap"
)

// acquireTokenWithRetry acquires token with retry and configurable timeout
// This is the new behavior enabled by feature flag
// Returns the access token or an error - caller is responsible for cancellation decisions
func (r *rbacClientImpl) acquireTokenWithRetry(ctx context.Context, reqID, internalReqID string, workspaceAttempt int, log *zap.SugaredLogger) (string, error) {
	var tokenResp *common.TokenResponse
	var tokenErr error

	// Retry token acquisition with shorter timeout per attempt
	for tokenAttempt := 0; tokenAttempt <= r.tokenMaxRetries; tokenAttempt++ {
		// Check if parent context was canceled
		if ctx.Err() != nil {
			if log != nil {
				log.Warnw("OIDC token acquisition aborted, upstream request canceled before attempt",
					"request_id", reqID,
					"internal_request_id", internalReqID,
					"token_attempt", tokenAttempt+1,
					"workspace_attempt", workspaceAttempt,
					"error", ctx.Err())
			}
			return "", ctx.Err() // Return raw error for clear error handling
		}

		tokenStart := time.Now()

		// Log token acquisition start
		if log != nil {
			log.Debugw("OIDC token acquisition started",
				"request_id", reqID,
				"internal_request_id", internalReqID,
				"token_attempt", tokenAttempt+1,
				"workspace_attempt", workspaceAttempt,
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
					"workspace_attempt", workspaceAttempt,
					"token_timeout_enabled", true)
			}
			return tokenResp.AccessToken, nil
		}

		// Token acquisition failed - distinguish error types
		// Check if parent context was canceled during token request
		// (This is NOT retryable - upstream caller gave up)
		if ctx.Err() != nil {
			if log != nil {
				log.Warnw("OIDC token acquisition failed, upstream context canceled during request",
					"request_id", reqID,
					"internal_request_id", internalReqID,
					"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
					"token_attempt", tokenAttempt+1,
					"workspace_attempt", workspaceAttempt,
					"upstream_error", ctx.Err(),
					"token_error", tokenErr,
					"token_timeout_enabled", true)
			}
			return "", ctx.Err() // Return raw error for clear error handling
		}

		// Token request failed (timeout, network, TLS, etc.) but parent context still valid
		if log != nil {
			log.Warnw("OIDC token acquisition failed",
				"request_id", reqID,
				"internal_request_id", internalReqID,
				"duration_seconds", fmt.Sprintf("%.3f", tokenDuration.Seconds()),
				"token_attempt", tokenAttempt+1,
				"workspace_attempt", workspaceAttempt,
				"error", tokenErr,
				"token_timeout_enabled", true)
		}

		// If this was the last retry, give up
		if tokenAttempt == r.tokenMaxRetries {
			// Categorize error for clearer alerting
			errorCategory := categorizeTokenError(tokenErr)
			return "", fmt.Errorf("token acquisition failed after %d attempts (%s): %w", r.tokenMaxRetries+1, errorCategory, tokenErr)
		}

		// Exponential backoff with jitter before retry
		// Uses same backoff strategy as workspace request retries
		// Context-aware sleep allows immediate cancellation during backoff
		backoff := r.calculateBackoff(tokenAttempt)
		select {
		case <-time.After(backoff):
			// Backoff complete, continue to next retry
		case <-ctx.Done():
			// Context canceled during backoff, exit immediately
			if log != nil {
				log.Warnw("OIDC token acquisition aborted, upstream context canceled during retry backoff",
					"request_id", reqID,
					"internal_request_id", internalReqID,
					"token_attempt", tokenAttempt+1,
					"workspace_attempt", workspaceAttempt,
					"backoff_ms", backoff.Milliseconds(),
					"error", ctx.Err())
			}
			return "", ctx.Err() // Return raw error for clear error handling
		}
	}

	// Should never reach here, but return error just in case
	return "", fmt.Errorf("token acquisition failed unexpectedly")
}

// categorizeTokenError inspects the error and returns a category for alerting
// Prefers typed error checking (errors.Is) over string matching for robustness
func categorizeTokenError(err error) string {
	if err == nil {
		return "unknown"
	}

	// Check well-known typed errors first (robust against wording/localization changes)
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	// Check for specific network error patterns via string matching
	// (these don't have typed sentinel errors)
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "TLS handshake timeout"):
		return "tls-timeout"
	case strings.Contains(errStr, "i/o timeout"):
		return "io-timeout"
	case strings.Contains(errStr, "connection refused"):
		return "connection-refused"
	case strings.Contains(errStr, "connection reset"):
		return "connection-reset"
	case strings.Contains(errStr, "no such host"):
		return "dns-error"
	}

	// Check if error is actually a network error (implements net.Error)
	// This catches real network errors without misclassifying auth/config/server errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return "network-error"
	}

	// Default to unknown for non-network errors (auth failures, config issues, etc.)
	return "unknown"
}
