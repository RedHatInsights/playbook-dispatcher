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
	"github.com/stretchr/testify/mock"
)

// mockTokenClient implements TokenClient interface for testing
type mockTokenClient struct {
	mock.Mock
}

func (m *mockTokenClient) GetToken() (*common.TokenResponse, error) {
	args := m.Called()
	if resp := args.Get(0); resp != nil {
		return resp.(*common.TokenResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTokenClient) GetTokenWithContext(ctx context.Context) (*common.TokenResponse, error) {
	args := m.Called(ctx)
	if resp := args.Get(0); resp != nil {
		return resp.(*common.TokenResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// TestAcquireTokenWithRetry_Success verifies successful token acquisition on first attempt
func TestAcquireTokenWithRetry_Success(t *testing.T) {
	mockToken := &mockTokenClient{}
	expectedResp := &common.TokenResponse{AccessToken: "test-token-12345"}

	mockToken.On("GetTokenWithContext", mock.Anything).
		Return(expectedResp, nil).Once()

	client := &rbacClientImpl{
		tokenClient:     mockToken,
		tokenMaxRetries: 2,
		tokenTimeout:    1 * time.Second,
	}

	token, err := client.acquireTokenWithRetry(context.Background(), "req-123", "internal-123", 1, nil)

	assert.NoError(t, err)
	assert.Equal(t, "test-token-12345", token)
	mockToken.AssertExpectations(t)
}

// TestAcquireTokenWithRetry_RetryOnce verifies retry logic on first failure
func TestAcquireTokenWithRetry_RetryOnce(t *testing.T) {
	mockToken := &mockTokenClient{}

	// First call fails with timeout
	mockToken.On("GetTokenWithContext", mock.Anything).
		Return(nil, context.DeadlineExceeded).Once()

	// Second call succeeds
	mockToken.On("GetTokenWithContext", mock.Anything).
		Return(&common.TokenResponse{AccessToken: "retry-token"}, nil).Once()

	client := &rbacClientImpl{
		tokenClient:     mockToken,
		tokenMaxRetries: 2,
		tokenTimeout:    1 * time.Second,
		initialBackoff:  1 * time.Millisecond, // Short backoff for testing
	}

	token, err := client.acquireTokenWithRetry(context.Background(), "req-123", "internal-123", 1, nil)

	assert.NoError(t, err)
	assert.Equal(t, "retry-token", token)
	mockToken.AssertExpectations(t)
	assert.Equal(t, 2, len(mockToken.Calls)) // Verify 2 attempts made
}

// TestAcquireTokenWithRetry_MaxRetriesExceeded verifies error after exhausting retries
func TestAcquireTokenWithRetry_MaxRetriesExceeded(t *testing.T) {
	mockToken := &mockTokenClient{}

	// All attempts fail
	mockToken.On("GetTokenWithContext", mock.Anything).
		Return(nil, errors.New("TLS handshake timeout")).Times(3) // maxRetries=2 means 3 total attempts

	client := &rbacClientImpl{
		tokenClient:     mockToken,
		tokenMaxRetries: 2,
		tokenTimeout:    1 * time.Second,
		initialBackoff:  1 * time.Millisecond,
	}

	token, err := client.acquireTokenWithRetry(context.Background(), "req-123", "internal-123", 1, nil)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "OIDC token acquisition failed after 3 attempts")
	assert.Contains(t, err.Error(), "tls-timeout") // Verify error categorization
	mockToken.AssertExpectations(t)
}

// TestAcquireTokenWithRetry_ContextCanceled verifies early exit on context cancellation
func TestAcquireTokenWithRetry_ContextCanceled(t *testing.T) {
	mockToken := &mockTokenClient{}
	// Mock should not be called since context is already canceled

	client := &rbacClientImpl{
		tokenClient:     mockToken,
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
	mockToken.AssertNotCalled(t, "GetTokenWithContext") // Verify no attempt was made
}

// TestAcquireTokenWithRetry_TimeoutEnforced verifies per-attempt timeout is enforced
func TestAcquireTokenWithRetry_TimeoutEnforced(t *testing.T) {
	mockToken := &mockTokenClient{}

	// Simulate slow token service that respects context
	mockToken.On("GetTokenWithContext", mock.Anything).
		Return(nil, context.DeadlineExceeded).Once()

	client := &rbacClientImpl{
		tokenClient:     mockToken,
		tokenMaxRetries: 0, // Single attempt
		tokenTimeout:    50 * time.Millisecond, // Very short timeout
	}

	start := time.Now()
	token, err := client.acquireTokenWithRetry(context.Background(), "req-123", "internal-123", 1, nil)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "timeout")
	// Verify we failed quickly due to timeout (allow some overhead)
	assert.Less(t, duration, 500*time.Millisecond)
	mockToken.AssertExpectations(t)
}

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
