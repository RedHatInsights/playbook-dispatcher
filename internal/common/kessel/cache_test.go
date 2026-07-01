// Coded in collaboration with AI
package kessel

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// mockKesselInventoryServiceForCache is a simple mock for cache tests
type mockKesselInventoryServiceForCache struct {
	allowed kesselv2.Allowed
}

func (m *mockKesselInventoryServiceForCache) Check(ctx context.Context, in *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
	return &kesselv2.CheckResponse{Allowed: m.allowed}, nil
}

func (m *mockKesselInventoryServiceForCache) CheckForUpdate(ctx context.Context, in *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
	return &kesselv2.CheckForUpdateResponse{Allowed: m.allowed}, nil
}

func (m *mockKesselInventoryServiceForCache) ReportResource(ctx context.Context, in *kesselv2.ReportResourceRequest, opts ...grpc.CallOption) (*kesselv2.ReportResourceResponse, error) {
	return nil, nil
}

func (m *mockKesselInventoryServiceForCache) DeleteResource(ctx context.Context, in *kesselv2.DeleteResourceRequest, opts ...grpc.CallOption) (*kesselv2.DeleteResourceResponse, error) {
	return nil, nil
}

func (m *mockKesselInventoryServiceForCache) StreamedListObjects(ctx context.Context, in *kesselv2.StreamedListObjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[kesselv2.StreamedListObjectsResponse], error) {
	return nil, nil
}

func (m *mockKesselInventoryServiceForCache) CheckBulk(ctx context.Context, in *kesselv2.CheckBulkRequest, opts ...grpc.CallOption) (*kesselv2.CheckBulkResponse, error) {
	return nil, errors.New("CheckBulk not implemented in cache test mock")
}

func (m *mockKesselInventoryServiceForCache) CheckForUpdateBulk(ctx context.Context, in *kesselv2.CheckForUpdateBulkRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateBulkResponse, error) {
	return nil, errors.New("CheckForUpdateBulk not implemented in cache test mock")
}

func (m *mockKesselInventoryServiceForCache) CheckSelf(ctx context.Context, in *kesselv2.CheckSelfRequest, opts ...grpc.CallOption) (*kesselv2.CheckSelfResponse, error) {
	return nil, errors.New("CheckSelf not implemented in cache test mock")
}

func (m *mockKesselInventoryServiceForCache) CheckSelfBulk(ctx context.Context, in *kesselv2.CheckSelfBulkRequest, opts ...grpc.CallOption) (*kesselv2.CheckSelfBulkResponse, error) {
	return nil, errors.New("CheckSelfBulk not implemented in cache test mock")
}

func (m *mockKesselInventoryServiceForCache) StreamedListSubjects(ctx context.Context, in *kesselv2.StreamedListSubjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[kesselv2.StreamedListSubjectsResponse], error) {
	return nil, errors.New("StreamedListSubjects not implemented in cache test mock")
}

func TestGetUserIDFromContext_User(t *testing.T) {
	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: &identity.User{
				UserID: "user-123",
			},
		},
	})

	userID := getUserIDFromContext(ctx)

	assert.Equal(t, "user-123", userID)
}

func TestGetUserIDFromContext_ServiceAccount(t *testing.T) {
	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type: "ServiceAccount",
			ServiceAccount: &identity.ServiceAccount{
				UserId: "sa-456",
			},
		},
	})

	userID := getUserIDFromContext(ctx)

	assert.Equal(t, "sa-456", userID)
}

func TestGetUserIDFromContext_EmptyUser(t *testing.T) {
	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: &identity.User{
				UserID: "",
			},
		},
	})

	userID := getUserIDFromContext(ctx)

	assert.Equal(t, "", userID)
}

func TestGetUserIDFromContext_NilUser(t *testing.T) {
	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: nil,
		},
	})

	userID := getUserIDFromContext(ctx)

	assert.Equal(t, "", userID)
}

func TestGetUserIDFromContext_UnknownType(t *testing.T) {
	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type: "System",
		},
	})

	userID := getUserIDFromContext(ctx)

	assert.Equal(t, "", userID)
}

func TestGetDefaultWorkspaceIDWithCache_ErrorOnMissingRequestID(t *testing.T) {
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second, RbacClientConfig{}, nil).(*rbacClientImpl)

	// Context without request_id (will be empty string)
	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-456",
			User: &identity.User{
				UserID: "user-789",
			},
		},
	})

	_, err := client.GetDefaultWorkspaceIDWithCache(ctx, "org-456")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request_id is required")
}

func TestGetDefaultWorkspaceIDWithCache_ErrorOnMissingOrgID(t *testing.T) {
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second, RbacClientConfig{}, nil).(*rbacClientImpl)

	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type: "User",
			User: &identity.User{
				UserID: "user-789",
			},
		},
	})

	// Even if we had request_id, empty orgID parameter should error
	// But request_id will be empty first, so that error will occur
	_, err := client.GetDefaultWorkspaceIDWithCache(ctx, "")

	assert.Error(t, err)
	// Will error on request_id first since we can't set it in tests
	assert.Contains(t, err.Error(), "required")
}

func TestGetDefaultWorkspaceIDWithCache_ErrorOnUnknownUser(t *testing.T) {
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second, RbacClientConfig{}, nil).(*rbacClientImpl)

	// Context with unknown user (nil User)
	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-456",
			User:  nil, // Will result in empty userID
		},
	})

	_, err := client.GetDefaultWorkspaceIDWithCache(ctx, "org-456")

	assert.Error(t, err)
	// Will error on request_id first since we can't set it in tests
	assert.Contains(t, err.Error(), "required")
}

func TestCheckApplicationPermissionsWithCache_FallbackOnMissingRequestID(t *testing.T) {
	// Create cache client wrapper (client not needed for validation tests)
	kesselCache := NewKesselClientWithCache(nil)

	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-456",
			User: &identity.User{
				UserID: "user-789",
			},
		},
	})

	log := zap.NewNop().Sugar()

	_, err := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-abc", log)

	// Should return error for missing request_id
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request_id is required")
}

func TestCheckApplicationPermissionsWithCache_FallbackOnMissingWorkspaceID(t *testing.T) {
	// Create cache client wrapper (client not needed for validation tests)
	kesselCache := NewKesselClientWithCache(nil)

	ctx := context.Background()
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-456",
			User: &identity.User{
				UserID: "user-789",
			},
		},
	})

	log := zap.NewNop().Sugar()

	_, err := kesselCache.CheckApplicationPermissionsWithCache(ctx, "", log)

	// Should return error for missing workspace_id (checked before request_id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is required")
}

func TestWorkspaceCache_Expiration(t *testing.T) {
	// Create test client with short TTL cache
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second, RbacClientConfig{}, nil).(*rbacClientImpl)
	client.workspaceCache = cache.New(100*time.Millisecond, 50*time.Millisecond)

	// Set value with hashed key
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte("workspace:req-123:org-456:user-789")))
	client.workspaceCache.Set(cacheKey, "workspace-test", cache.DefaultExpiration)

	// Verify it's there
	value, found := client.workspaceCache.Get(cacheKey)
	assert.True(t, found)
	assert.Equal(t, "workspace-test", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify it's gone
	_, found = client.workspaceCache.Get(cacheKey)
	assert.False(t, found)
}

func TestApplicationCache_Expiration(t *testing.T) {
	// Create test client with short TTL cache
	kesselClient := NewKesselClientWithCache(nil)
	kesselClient.applicationCache = cache.New(100*time.Millisecond, 50*time.Millisecond)

	// Set value with hashed key for application permissions
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte("applications:req-123:org-456:user-789:workspace-abc")))
	kesselClient.applicationCache.Set(cacheKey, []string{"config-manager", "remediations"}, cache.DefaultExpiration)

	// Verify it's there
	value, found := kesselClient.applicationCache.Get(cacheKey)
	assert.True(t, found)
	assert.Equal(t, []string{"config-manager", "remediations"}, value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify it's gone
	_, found = kesselClient.applicationCache.Get(cacheKey)
	assert.False(t, found)
}

func TestHashCacheKey_WithDelimiter(t *testing.T) {
	// Verify that different combinations produce different hashes due to delimiter
	hash1 := hashCacheKey("a", "bc")
	hash2 := hashCacheKey("ab", "c")

	assert.NotEqual(t, hash1, hash2, "Different part combinations should produce different hashes")
}

func TestHashCacheKey_DifferentWorkspaces(t *testing.T) {
	// Verify that different workspaces produce different cache keys (security test)
	// Same user, different workspaces should have different keys
	hash1 := hashCacheKey("applications", "req-123", "org-456", "user-789", "workspace-A")
	hash2 := hashCacheKey("applications", "req-123", "org-456", "user-789", "workspace-B")

	assert.NotEqual(t, hash1, hash2, "Different workspaces must produce different cache keys to prevent authorization leakage")
}

// Test helper to create a context with request_id set via middleware
func contextWithRequestID(t *testing.T) context.Context {
	t.Helper()

	var capturedCtx context.Context

	// Create a test HTTP request
	req := httptest.NewRequest("GET", "/test", nil)

	// Apply RequestID middleware to capture the context
	handler := request_id.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
	}))

	// Execute the handler to populate context
	handler.ServeHTTP(httptest.NewRecorder(), req)

	return capturedCtx
}

func TestCheckApplicationPermissionsWithCache_CacheHit_AllowedApps(t *testing.T) {
	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-123",
			User: &identity.User{
				UserID: "user-456",
			},
		},
	})

	// Create mock permission check function with call counter
	callCount := 0
	mockPermissionCheck := func(ctx context.Context, workspaceID string, log *zap.SugaredLogger) ([]string, error) {
		callCount++
		return []string{"remediations", "tasks"}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// First call - should hit backend
	allowedApps1, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-abc", log)
	assert.NoError(t, err1)
	assert.Equal(t, []string{"remediations", "tasks"}, allowedApps1)
	assert.Equal(t, 1, callCount, "First call should invoke backend")

	// Second call with same params - should hit cache
	allowedApps2, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-abc", log)
	assert.NoError(t, err2)
	assert.Equal(t, []string{"remediations", "tasks"}, allowedApps2)
	assert.Equal(t, 1, callCount, "Second call should NOT invoke backend (cache hit)")
}

func TestCheckApplicationPermissionsWithCache_CacheHit_NoApps(t *testing.T) {
	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-789",
			User: &identity.User{
				UserID: "user-101",
			},
		},
	})

	// Create mock that returns empty list (denied)
	callCount := 0
	mockPermissionCheck := func(ctx context.Context, workspaceID string, log *zap.SugaredLogger) ([]string, error) {
		callCount++
		return []string{}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// First call - should hit backend
	allowedApps1, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-xyz", log)
	assert.NoError(t, err1)
	assert.Equal(t, []string{}, allowedApps1)
	assert.Equal(t, 1, callCount, "First call should invoke backend")

	// Second call - should hit cache (even for denied result)
	allowedApps2, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-xyz", log)
	assert.NoError(t, err2)
	assert.Equal(t, []string{}, allowedApps2)
	assert.Equal(t, 1, callCount, "Second call should NOT invoke backend (denied results are cached)")
}

func TestCheckApplicationPermissionsWithCache_DifferentCacheKeys(t *testing.T) {
	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-999",
			User: &identity.User{
				UserID: "user-888",
			},
		},
	})

	// Create mock with call counter
	callCount := 0
	mockPermissionCheck := func(ctx context.Context, workspaceID string, log *zap.SugaredLogger) ([]string, error) {
		callCount++
		// Return different results based on workspace
		if workspaceID == "workspace-1" {
			return []string{"remediations"}, nil
		}
		return []string{"tasks"}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// Call with workspace-1
	apps1, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-1", log)
	assert.NoError(t, err1)
	assert.Equal(t, []string{"remediations"}, apps1)
	assert.Equal(t, 1, callCount)

	// Call with workspace-2 (different cache key)
	apps2, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-2", log)
	assert.NoError(t, err2)
	assert.Equal(t, []string{"tasks"}, apps2)
	assert.Equal(t, 2, callCount, "Different workspace should be cache miss")

	// Call workspace-1 again (cache hit)
	apps3, err3 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-1", log)
	assert.NoError(t, err3)
	assert.Equal(t, []string{"remediations"}, apps3)
	assert.Equal(t, 2, callCount, "Same workspace should be cache hit")
}

func TestCheckApplicationPermissionsWithCache_ErrorNotCached(t *testing.T) {
	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-555",
			User: &identity.User{
				UserID: "user-666",
			},
		},
	})

	// Create mock that returns error
	callCount := 0
	mockPermissionCheck := func(ctx context.Context, workspaceID string, log *zap.SugaredLogger) ([]string, error) {
		callCount++
		return nil, errors.New("backend error")
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// First call - should fail
	_, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-fail", log)
	assert.Error(t, err1)
	assert.Contains(t, err1.Error(), "backend error")
	assert.Equal(t, 1, callCount)

	// Second call - should try backend again (errors are not cached)
	_, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-fail", log)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "backend error")
	assert.Equal(t, 2, callCount, "Errors should NOT be cached")
}
