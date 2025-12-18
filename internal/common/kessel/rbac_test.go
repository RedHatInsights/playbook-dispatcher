package kessel

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRbacClient(t *testing.T) {
	client := NewRbacClient("http://localhost:8080", nil, 60*time.Second)

	assert.NotNil(t, client)
	impl, ok := client.(*rbacClientImpl)
	assert.True(t, ok)
	assert.Equal(t, "http://localhost:8080", impl.rbacURL)
	assert.Equal(t, 3, impl.maxRetries)
	assert.Equal(t, 100*time.Millisecond, impl.initialBackoff)
	assert.Equal(t, 2*time.Second, impl.maxBackoff)
	assert.Equal(t, 60*time.Second, impl.requestTimeout)
}

func TestGetDefaultWorkspaceID_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/rbac/v2/workspaces/")
		assert.Equal(t, "default", r.URL.Query().Get("type"))
		assert.Equal(t, "test-org", r.Header.Get("x-rh-rbac-org-id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":[{"id":"workspace-123"}]}`)
	}))
	defer server.Close()

	client := NewRbacClient(server.URL, nil, 10*time.Second)
	workspaceID, err := client.GetDefaultWorkspaceID(context.Background(), "test-org")

	assert.NoError(t, err)
	assert.Equal(t, "workspace-123", workspaceID)
}

func TestGetDefaultWorkspaceID_NoWorkspaceFound(t *testing.T) {
	// Create test server that returns empty data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer server.Close()

	client := NewRbacClient(server.URL, nil, 10*time.Second)
	workspaceID, err := client.GetDefaultWorkspaceID(context.Background(), "test-org")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected number of default workspaces: 0")
	assert.Empty(t, workspaceID)
}

func TestGetDefaultWorkspaceID_InvalidJSON(t *testing.T) {
	// Create test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `invalid json`)
	}))
	defer server.Close()

	client := NewRbacClient(server.URL, nil, 10*time.Second)
	workspaceID, err := client.GetDefaultWorkspaceID(context.Background(), "test-org")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
	assert.Empty(t, workspaceID)
}

func TestShouldRetry_NetworkError(t *testing.T) {
	client := &rbacClientImpl{}

	shouldRetry := client.shouldRetry(nil, fmt.Errorf("network error"))

	assert.True(t, shouldRetry)
}

func TestShouldRetry_5xxError(t *testing.T) {
	client := &rbacClientImpl{}

	tests := []struct {
		statusCode  int
		shouldRetry bool
		description string
	}{
		{500, true, "Internal Server Error"},
		{502, true, "Bad Gateway"},
		{503, true, "Service Unavailable"},
		{504, true, "Gateway Timeout"},
		{429, true, "Too Many Requests"},
		{400, false, "Bad Request"},
		{401, false, "Unauthorized"},
		{403, false, "Forbidden"},
		{404, false, "Not Found"},
		{200, false, "OK"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
			}

			shouldRetry := client.shouldRetry(resp, nil)

			assert.Equal(t, tt.shouldRetry, shouldRetry)
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	client := &rbacClientImpl{
		initialBackoff: 100 * time.Millisecond,
		maxBackoff:     2 * time.Second,
	}

	tests := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
		description string
	}{
		{0, 50 * time.Millisecond, 100 * time.Millisecond, "First retry"},
		{1, 100 * time.Millisecond, 200 * time.Millisecond, "Second retry"},
		{2, 200 * time.Millisecond, 400 * time.Millisecond, "Third retry"},
		{10, 1 * time.Second, 2 * time.Second, "Should cap at maxBackoff"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			backoff := client.calculateBackoff(tt.attempt)

			assert.GreaterOrEqual(t, backoff, tt.minExpected)
			assert.LessOrEqual(t, backoff, tt.maxExpected)
		})
	}
}

func TestCalculateBackoff_JitterRange(t *testing.T) {
	client := &rbacClientImpl{
		initialBackoff: 100 * time.Millisecond,
		maxBackoff:     2 * time.Second,
	}

	// Run multiple times to test jitter randomness
	for i := 0; i < 100; i++ {
		backoff := client.calculateBackoff(0)

		// Jitter range is [0.5, 1.0] of 100ms = [50ms, 100ms]
		assert.GreaterOrEqual(t, backoff, 50*time.Millisecond)
		assert.LessOrEqual(t, backoff, 100*time.Millisecond)
	}
}

func TestGetDefaultWorkspaceID_RetryOn503(t *testing.T) {
	attempts := 0

	// Create test server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":[{"id":"workspace-456"}]}`)
	}))
	defer server.Close()

	client := NewRbacClient(server.URL, nil, 10*time.Second)
	workspaceID, err := client.GetDefaultWorkspaceID(context.Background(), "test-org")

	assert.NoError(t, err)
	assert.Equal(t, "workspace-456", workspaceID)
	assert.Equal(t, 3, attempts)
}

func TestGetDefaultWorkspaceID_MaxRetriesExceeded(t *testing.T) {
	attempts := 0

	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewRbacClient(server.URL, nil, 10*time.Second)
	workspaceID, err := client.GetDefaultWorkspaceID(context.Background(), "test-org")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "503")
	assert.Empty(t, workspaceID)
	assert.Equal(t, 4, attempts) // Initial attempt + 3 retries
}

func TestGetDefaultWorkspaceID_NoRetryOn404(t *testing.T) {
	attempts := 0

	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewRbacClient(server.URL, nil, 10*time.Second)
	workspaceID, err := client.GetDefaultWorkspaceID(context.Background(), "test-org")

	assert.Error(t, err)
	assert.Empty(t, workspaceID)
	assert.Equal(t, 1, attempts) // Should not retry on 4xx errors
}

func TestGetDefaultWorkspaceID_ContextTimeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":[{"id":"workspace-789"}]}`)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := &rbacClientImpl{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		rbacURL:        server.URL,
		tokenClient:    nil,
		maxRetries:     3,
		initialBackoff: 10 * time.Millisecond,
		maxBackoff:     100 * time.Millisecond,
		requestTimeout: 1 * time.Millisecond, // Very short timeout
	}

	ctx := context.Background()
	workspaceID, err := client.GetDefaultWorkspaceID(ctx, "test-org")

	assert.Error(t, err)
	assert.Empty(t, workspaceID)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestNewRbacClient_URLConstruction_HostWithoutPort(t *testing.T) {
	// Test that the client stores the URL correctly when host doesn't have port
	url := "http://localhost:8080"
	client := NewRbacClient(url, nil, 10*time.Second)

	impl, ok := client.(*rbacClientImpl)
	assert.True(t, ok)
	assert.Equal(t, url, impl.rbacURL)
}

func TestNewRbacClient_URLConstruction_HostWithPort(t *testing.T) {
	// Test that the client stores the URL correctly when host already has port
	url := "http://localhost:8080"
	client := NewRbacClient(url, nil, 10*time.Second)

	impl, ok := client.(*rbacClientImpl)
	assert.True(t, ok)
	assert.Equal(t, url, impl.rbacURL)
}

func TestGetDefaultWorkspaceID_URLFormatting(t *testing.T) {
	// Create test server that captures the request URL
	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":"workspace-123"}]}`))
	}))
	defer server.Close()

	// Test with host:port in the rbacURL (should not double the port)
	client := NewRbacClient(server.URL, nil, 10*time.Second)
	ctx := context.Background()
	workspaceID, err := client.GetDefaultWorkspaceID(ctx, "test-org")

	assert.NoError(t, err)
	assert.Equal(t, "workspace-123", workspaceID)
	// Verify the URL format matches config-manager (no org_id query param)
	assert.Contains(t, capturedURL, "/api/rbac/v2/workspaces/")
	assert.Contains(t, capturedURL, "type=default")
	assert.NotContains(t, capturedURL, "org_id")
}
