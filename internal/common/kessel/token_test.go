package kessel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/project-kessel/inventory-client-go/common"
	"github.com/stretchr/testify/assert"
)

// TestAcquireTokenWithRetry_ContextCanceled verifies early exit on context cancellation
func TestAcquireTokenWithRetry_ContextCanceled(t *testing.T) {
	client := &rbacClientImpl{
		tokenClient:     &common.TokenClient{}, // Mock client
		tokenMaxRetries: 2,
		tokenTimeout:    1 * time.Second,
	}

	// Create already-canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	token, err := client.acquireTokenWithRetry(ctx, "req-123", "internal-123", 1, nil)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err, "Should return raw context.Canceled error")
	assert.Empty(t, token)
}

// Note: Full integration testing of token retry logic with actual token client
// and feature flag branching is covered by existing tests in rbac_test.go
// This test only verifies the retry loop's early exit on pre-canceled context

func TestCategorizeTokenError(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedCategory string
	}{
		{
			name:             "nil error returns unknown",
			err:              nil,
			expectedCategory: "unknown",
		},
		{
			name:             "typed context.Canceled error",
			err:              context.Canceled,
			expectedCategory: "canceled",
		},
		{
			name:             "wrapped context.Canceled error",
			err:              fmt.Errorf("token request failed: %w", context.Canceled),
			expectedCategory: "canceled",
		},
		{
			name:             "typed context.DeadlineExceeded error",
			err:              context.DeadlineExceeded,
			expectedCategory: "timeout",
		},
		{
			name:             "wrapped context.DeadlineExceeded error",
			err:              fmt.Errorf("Post \"https://sso.redhat.com\": %w", context.DeadlineExceeded),
			expectedCategory: "timeout",
		},
		{
			name:             "TLS handshake timeout (no typed error, uses string match)",
			err:              errors.New("Post \"https://sso.redhat.com\": net/http: TLS handshake timeout"),
			expectedCategory: "tls-timeout",
		},
		{
			name:             "i/o timeout (no typed error, uses string match)",
			err:              errors.New("dial tcp: i/o timeout"),
			expectedCategory: "io-timeout",
		},
		{
			name:             "connection refused (no typed error, uses string match)",
			err:              errors.New("dial tcp: connection refused"),
			expectedCategory: "connection-refused",
		},
		{
			name:             "connection reset (no typed error, uses string match)",
			err:              errors.New("read tcp: connection reset by peer"),
			expectedCategory: "connection-reset",
		},
		{
			name:             "DNS error (no typed error, uses string match)",
			err:              errors.New("dial tcp: lookup sso.redhat.com: no such host"),
			expectedCategory: "dns-error",
		},
		{
			name:             "generic error without pattern match returns unknown",
			err:              errors.New("some random error"),
			expectedCategory: "unknown",
		},
		{
			name:             "actual network error (implements net.Error)",
			err:              &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection failed")},
			expectedCategory: "network-error",
		},
		{
			name:             "TLS timeout detected via string when not wrapped",
			err:              errors.New("TLS handshake timeout"),
			expectedCategory: "tls-timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := categorizeTokenError(tt.err)
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}
