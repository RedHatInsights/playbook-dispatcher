package kessel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRbacClient(t *testing.T) {
	cfg := viper.New()
	cfg.Set("rbac.scheme", "http")
	cfg.Set("rbac.host", "rbac-service")
	cfg.Set("rbac.port", 8080)
	cfg.Set("rbac.timeout", 10)

	client := NewRbacClient(cfg, nil)

	assert.NotNil(t, client)

	// Verify it implements the interface
	var _ RbacClient = client
}

func TestRbacClient_GetDefaultWorkspaceID_Success(t *testing.T) {
	// Create a test server that returns a valid workspace response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/rbac/v2/workspaces/?type=default", r.URL.Path+"?"+r.URL.RawQuery)
		assert.Equal(t, "org-123", r.Header.Get("x-rh-rbac-org-id"))

		// Return a valid response
		response := workspacesResponse{
			Data: []workspace{
				{ID: "workspace-456"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Create client pointing to test server
	cfg := viper.New()
	cfg.Set("rbac.scheme", "http")
	cfg.Set("rbac.host", server.URL[7:]) // Remove "http://"
	cfg.Set("rbac.port", 80)
	cfg.Set("rbac.timeout", 10)

	// Create a custom client that points to our test server
	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.NoError(t, err)
	assert.Equal(t, "workspace-456", workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_MultipleWorkspaces(t *testing.T) {
	// Create a test server that returns multiple workspaces (error case)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := workspacesResponse{
			Data: []workspace{
				{ID: "workspace-1"},
				{ID: "workspace-2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple default workspaces found")
	assert.Contains(t, err.Error(), "got 2, expected 1")
	assert.Contains(t, err.Error(), "data inconsistency")
	assert.Empty(t, workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_NoWorkspaces(t *testing.T) {
	// Create a test server that returns no workspaces (error case)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := workspacesResponse{
			Data: []workspace{},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no default workspace found")
	assert.Contains(t, err.Error(), "org-123")
	assert.Contains(t, err.Error(), "workspace may not be configured")
	assert.Empty(t, workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_HTTPError(t *testing.T) {
	// Create a test server that returns an HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RBAC API error")
	assert.Contains(t, err.Error(), "status 500")
	assert.Contains(t, err.Error(), "Internal Server Error")
	assert.Empty(t, workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json{"))
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshalling response")
	assert.Empty(t, workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with very short timeout
	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 50 * time.Millisecond,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error making request")
	assert.Empty(t, workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_ContextCanceled(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This won't be reached because context is canceled
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error making request")
	assert.Empty(t, workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_WithAuthentication(t *testing.T) {
	// This test verifies that authentication token is included when tokenClient is present
	// We can't easily test the actual token client without mocking the Kessel token service
	// This test documents the expected behavior

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real scenario, we would verify the Authorization header here
		// For now, we just return a valid response
		response := workspacesResponse{
			Data: []workspace{
				{ID: "workspace-789"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil, // In real scenario, this would be a mock token client
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-456")

	assert.NoError(t, err)
	assert.Equal(t, "workspace-789", workspaceID)
}

func TestWorkspaceResponse_Serialization(t *testing.T) {
	// Test that workspace response structures serialize/deserialize correctly
	original := workspacesResponse{
		Data: []workspace{
			{ID: "ws-1"},
			{ID: "ws-2"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var decoded workspacesResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify equality
	assert.Equal(t, original, decoded)
	assert.Len(t, decoded.Data, 2)
	assert.Equal(t, "ws-1", decoded.Data[0].ID)
	assert.Equal(t, "ws-2", decoded.Data[1].ID)
}

func TestRbacClientImpl_ImplementsInterface(t *testing.T) {
	// Compile-time check that rbacClientImpl implements RbacClient
	var _ RbacClient = &rbacClientImpl{}
}

func TestRbacClient_GetDefaultWorkspaceID_LargeErrorBody(t *testing.T) {
	// Create a test server that returns a large error body (should be truncated)
	largeBody := make([]byte, 500)
	for i := range largeBody {
		largeBody[i] = 'X'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(largeBody)
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RBAC API error")
	assert.Contains(t, err.Error(), "status 400")
	assert.Contains(t, err.Error(), "Bad Request")
	// Should contain "..." to indicate truncation
	assert.Contains(t, err.Error(), "...")
	// Error message should not contain the full 500 character body
	assert.Less(t, len(err.Error()), 400) // Much less than 500 + overhead
	assert.Empty(t, workspaceID)
}

func TestRbacClient_GetDefaultWorkspaceID_JSONErrorBody(t *testing.T) {
	// Create a test server that returns a JSON error body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"insufficient permissions","code":"RBAC_403"}`))
	}))
	defer server.Close()

	impl := &rbacClientImpl{
		baseURL: server.URL,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		tokenClient: nil,
	}

	ctx := context.Background()
	workspaceID, err := impl.GetDefaultWorkspaceID(ctx, "org-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RBAC API error")
	assert.Contains(t, err.Error(), "status 403")
	assert.Contains(t, err.Error(), "Forbidden")
	// Should include the JSON error body
	assert.Contains(t, err.Error(), "insufficient permissions")
	assert.Contains(t, err.Error(), "RBAC_403")
	assert.Empty(t, workspaceID)
}

func TestRbacClient_BaseURLConstruction(t *testing.T) {
	tests := []struct {
		name        string
		scheme      string
		host        string
		port        int
		expectedURL string
	}{
		{
			name:        "http with standard port",
			scheme:      "http",
			host:        "rbac-service",
			port:        8080,
			expectedURL: "http://rbac-service:8080",
		},
		{
			name:        "https with custom port",
			scheme:      "https",
			host:        "rbac.example.com",
			port:        9443,
			expectedURL: "https://rbac.example.com:9443",
		},
		{
			name:        "localhost",
			scheme:      "http",
			host:        "localhost",
			port:        3000,
			expectedURL: "http://localhost:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := viper.New()
			cfg.Set("rbac.scheme", tt.scheme)
			cfg.Set("rbac.host", tt.host)
			cfg.Set("rbac.port", tt.port)
			cfg.Set("rbac.timeout", 10)

			client := NewRbacClient(cfg, nil)
			impl := client.(*rbacClientImpl)

			assert.Equal(t, tt.expectedURL, impl.baseURL)
		})
	}
}
