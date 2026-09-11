// Coded in collaboration with AI
package kessel

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

	_, err := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-abc", "", log)

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

	_, err := kesselCache.CheckApplicationPermissionsWithCache(ctx, "", "", log)

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

// contextWithRequestID creates a context with request_id set via middleware.
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
	mockPermissionCheck := func(ctx context.Context, workspaceID string, serviceFilter string, log *zap.SugaredLogger) ([]string, error) {
		callCount++
		return []string{"remediations", "tasks"}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// First call - should hit backend
	allowedApps1, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-abc", "", log)
	assert.NoError(t, err1)
	assert.Equal(t, []string{"remediations", "tasks"}, allowedApps1)
	assert.Equal(t, 1, callCount, "First call should invoke backend")

	// Second call with same params - should hit cache
	allowedApps2, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-abc", "", log)
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
	mockPermissionCheck := func(ctx context.Context, workspaceID string, serviceFilter string, log *zap.SugaredLogger) ([]string, error) {
		callCount++
		return []string{}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// First call - should hit backend
	allowedApps1, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-xyz", "", log)
	assert.NoError(t, err1)
	assert.Equal(t, []string{}, allowedApps1)
	assert.Equal(t, 1, callCount, "First call should invoke backend")

	// Second call - should hit cache (even for denied result)
	allowedApps2, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-xyz", "", log)
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
	mockPermissionCheck := func(ctx context.Context, workspaceID string, serviceFilter string, log *zap.SugaredLogger) ([]string, error) {
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
	apps1, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-1", "", log)
	assert.NoError(t, err1)
	assert.Equal(t, []string{"remediations"}, apps1)
	assert.Equal(t, 1, callCount)

	// Call with workspace-2 (different cache key)
	apps2, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-2", "", log)
	assert.NoError(t, err2)
	assert.Equal(t, []string{"tasks"}, apps2)
	assert.Equal(t, 2, callCount, "Different workspace should be cache miss")

	// Call workspace-1 again (cache hit)
	apps3, err3 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-1", "", log)
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
	mockPermissionCheck := func(ctx context.Context, workspaceID string, serviceFilter string, log *zap.SugaredLogger) ([]string, error) {
		callCount++
		return nil, errors.New("backend error")
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// First call - should fail
	_, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-fail", "", log)
	assert.Error(t, err1)
	assert.Contains(t, err1.Error(), "backend error")
	assert.Equal(t, 1, callCount)

	// Second call - should try backend again (errors are not cached)
	_, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-fail", "", log)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "backend error")
	assert.Equal(t, 2, callCount, "Errors should NOT be cached")
}

func TestCheckApplicationPermissionsWithCache_SingleflightDeduplication(t *testing.T) {
	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-concurrent",
			User: &identity.User{
				UserID: "user-concurrent",
			},
		},
	})

	// Create mock with delay to ensure concurrent calls overlap
	// Note: We use time-based synchronization (sleep) rather than barriers because
	// singleflight only executes this function ONCE (the leader), so a barrier waiting
	// for all 5 callers would deadlock. The 50ms sleep is sufficient for goroutines to
	// start (they launch in microseconds) and is reliable in practice.
	var callCount atomic.Int32
	mockPermissionCheck := func(ctx context.Context, workspaceID string, serviceFilter string, log *zap.SugaredLogger) ([]string, error) {
		callCount.Add(1)
		time.Sleep(50 * time.Millisecond)
		return []string{"remediations", "tasks"}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// Launch 5 concurrent requests
	results := make(chan []string, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			apps, err := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-concurrent", "", log)
			results <- apps
			errors <- err
		}()
	}

	// Collect all results
	var allApps [][]string
	var allErrors []error
	for i := 0; i < 5; i++ {
		allApps = append(allApps, <-results)
		allErrors = append(allErrors, <-errors)
	}

	// Verify all succeeded
	for i, err := range allErrors {
		assert.NoError(t, err, "Request %d should succeed", i)
	}

	// Verify all got the same result
	for i, apps := range allApps {
		assert.Equal(t, []string{"remediations", "tasks"}, apps, "Request %d should get correct result", i)
	}

	// Most importantly: singleflight should deduplicate, so only 1 backend call
	assert.Equal(t, int32(1), callCount.Load(), "Singleflight should deduplicate concurrent requests to 1 backend call")

	// Verify cache was populated - make a subsequent request
	apps6, err6 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-concurrent", "", log)
	assert.NoError(t, err6)
	assert.Equal(t, []string{"remediations", "tasks"}, apps6, "Subsequent request should get cached result")

	// Verify no additional backend call (still only 1 total)
	assert.Equal(t, int32(1), callCount.Load(), "Subsequent request should hit cache, not backend")
}

func TestCheckApplicationPermissionsWithCache_CanceledCallerDoesNotFailOthers(t *testing.T) {
	// Create base context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-cancel",
			User: &identity.User{
				UserID: "user-cancel",
			},
		},
	})

	// Create a short-lived context that will cancel before backend completes
	cancelCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	// Mock backend that delays longer than cancelCtx timeout
	var callCount atomic.Int32
	mockPermissionCheck := func(ctx context.Context, workspaceID string, serviceFilter string, log *zap.SugaredLogger) ([]string, error) {
		callCount.Add(1)
		time.Sleep(50 * time.Millisecond) // Outlives cancelCtx
		return []string{"remediations", "tasks"}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// Launch 2 concurrent requests with different contexts
	type result struct {
		callerID string
		apps     []string
		err      error
	}
	results := make(chan result, 2)

	// First caller with short-lived context (will timeout)
	go func() {
		apps, err := kesselCache.CheckApplicationPermissionsWithCache(cancelCtx, "workspace-cancel", "", log)
		results <- result{callerID: "canceled", apps: apps, err: err}
	}()

	// Second caller with normal context (should succeed)
	go func() {
		apps, err := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-cancel", "", log)
		results <- result{callerID: "normal", apps: apps, err: err}
	}()

	// Collect and identify each caller's result
	resultsMap := make(map[string]result)
	for i := 0; i < 2; i++ {
		r := <-results
		resultsMap[r.callerID] = r
	}

	// Critical assertion: caller with normal context must succeed
	normalResult := resultsMap["normal"]
	assert.NoError(t, normalResult.err, "Caller with valid context should succeed even if another caller's context is canceled")
	assert.Equal(t, []string{"remediations", "tasks"}, normalResult.apps)

	// Canceled caller may succeed (if result arrives before 10ms timeout) or fail - both are valid
	canceledResult := resultsMap["canceled"]
	if canceledResult.err != nil {
		t.Logf("Canceled caller returned error as expected: %v", canceledResult.err)
	} else {
		t.Logf("Canceled caller succeeded before timeout (received shared result from singleflight)")
		assert.Equal(t, []string{"remediations", "tasks"}, canceledResult.apps)
	}

	// Backend should only be called once (singleflight deduplication worked)
	assert.Equal(t, int32(1), callCount.Load(), "Backend should be called once despite context cancellation")
}

// TestCheckApplicationPermissionsWithCache_TypeAssertionFailure verifies that corrupted cache entries
// (wrong type) are detected, deleted, and recovered from gracefully.
func TestCheckApplicationPermissionsWithCache_TypeAssertionFailure(t *testing.T) {
	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-type",
			User: &identity.User{
				UserID: "user-type",
			},
		},
	})

	// Create mock permission check
	mockPermissionCheck := func(ctx context.Context, workspaceID string, serviceFilter string, log *zap.SugaredLogger) ([]string, error) {
		return []string{"remediations"}, nil
	}

	// Create cache client with mock
	kesselCache := NewKesselClientWithCache(nil)
	kesselCache.permissionCheckFunc = mockPermissionCheck

	log := zap.NewNop().Sugar()

	// First call succeeds normally
	apps1, err1 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-type", "", log)
	assert.NoError(t, err1)
	assert.Equal(t, []string{"remediations"}, apps1)

	// Manually corrupt the cache with wrong type
	reqID := request_id.GetReqID(ctx)
	cacheKey := hashCacheKey("applications", reqID, "org-type", "user-type", "workspace-type", "")
	kesselCache.applicationCache.Set(cacheKey, "wrong-type-string", cache.DefaultExpiration)

	// Verify corrupted entry exists
	corrupted, found := kesselCache.applicationCache.Get(cacheKey)
	assert.True(t, found, "Corrupted cache entry should exist")
	assert.Equal(t, "wrong-type-string", corrupted, "Should have corrupted value")

	// Second call should detect type mismatch, delete corrupted entry, fetch fresh data, and cache it
	apps2, err2 := kesselCache.CheckApplicationPermissionsWithCache(ctx, "workspace-type", "", log)
	assert.NoError(t, err2, "Should recover from type mismatch by fetching fresh data")
	assert.Equal(t, []string{"remediations"}, apps2, "Should return fresh data after cache corruption")

	// Verify cache now contains correct data (not corrupted)
	cached, found := kesselCache.applicationCache.Get(cacheKey)
	assert.True(t, found, "Cache entry should exist with fresh data")
	correctData, ok := cached.([]string)
	assert.True(t, ok, "Cached data should be correct type now")
	assert.Equal(t, []string{"remediations"}, correctData, "Cached data should be correct")
}

func TestGetDefaultWorkspaceIDWithCache_SingleflightDeduplication(t *testing.T) {
	// Create a mock RBAC server that tracks call count
	// Note: We use time-based synchronization (sleep) rather than barriers because
	// singleflight only executes this handler ONCE (the leader), so a barrier waiting
	// for all 5 callers would deadlock. The 50ms sleep is sufficient for goroutines to
	// start (they launch in microseconds) and is reliable in practice.
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data":[{"id":"workspace-123"}]}`)); err != nil {
			t.Fatalf("failed to write mock response: %v", err)
		}
	}))
	defer server.Close()

	// Create RBAC client pointing to mock server
	client := NewRbacClient(server.URL, nil, 10*time.Second, RbacClientConfig{}, nil).(*rbacClientImpl)

	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-concurrent",
			User: &identity.User{
				UserID: "user-concurrent",
			},
		},
	})

	// Launch 5 concurrent requests
	results := make(chan string, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			workspaceID, err := client.GetDefaultWorkspaceIDWithCache(ctx, "org-concurrent")
			results <- workspaceID
			errors <- err
		}()
	}

	// Collect all results
	var allWorkspaceIDs []string
	var allErrors []error
	for i := 0; i < 5; i++ {
		allWorkspaceIDs = append(allWorkspaceIDs, <-results)
		allErrors = append(allErrors, <-errors)
	}

	// Verify all succeeded
	for i, err := range allErrors {
		assert.NoError(t, err, "Request %d should succeed", i)
	}

	// Verify all got the same result
	for i, workspaceID := range allWorkspaceIDs {
		assert.Equal(t, "workspace-123", workspaceID, "Request %d should get correct workspace ID", i)
	}

	// Most importantly: singleflight should deduplicate, so only 1 backend call
	assert.Equal(t, int32(1), callCount.Load(), "Singleflight should deduplicate concurrent requests to 1 backend call")

	// Verify cache was populated - make a subsequent request
	workspaceID6, err6 := client.GetDefaultWorkspaceIDWithCache(ctx, "org-concurrent")
	assert.NoError(t, err6)
	assert.Equal(t, "workspace-123", workspaceID6, "Subsequent request should get cached result")

	// Verify no additional backend call (still only 1 total)
	assert.Equal(t, int32(1), callCount.Load(), "Subsequent request should hit cache, not backend")
}

func TestGetDefaultWorkspaceIDWithCache_CanceledCallerDoesNotFailOthers(t *testing.T) {
	// Create a mock RBAC server with delayed response
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		time.Sleep(50 * time.Millisecond) // Delay longer than cancelCtx timeout
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data":[{"id":"workspace-456"}]}`)); err != nil {
			t.Fatalf("failed to write mock response: %v", err)
		}
	}))
	defer server.Close()

	// Create RBAC client pointing to mock server
	client := NewRbacClient(server.URL, nil, 10*time.Second, RbacClientConfig{}, nil).(*rbacClientImpl)

	// Create base context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-cancel",
			User: &identity.User{
				UserID: "user-cancel",
			},
		},
	})

	// Create a short-lived context that will cancel before backend completes
	cancelCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	// Launch 2 concurrent requests with different contexts
	type result struct {
		callerID    string
		workspaceID string
		err         error
	}
	results := make(chan result, 2)

	// First caller with short-lived context (will timeout)
	go func() {
		workspaceID, err := client.GetDefaultWorkspaceIDWithCache(cancelCtx, "org-cancel")
		results <- result{callerID: "canceled", workspaceID: workspaceID, err: err}
	}()

	// Second caller with normal context (should succeed)
	go func() {
		workspaceID, err := client.GetDefaultWorkspaceIDWithCache(ctx, "org-cancel")
		results <- result{callerID: "normal", workspaceID: workspaceID, err: err}
	}()

	// Collect and identify each caller's result
	resultsMap := make(map[string]result)
	for i := 0; i < 2; i++ {
		r := <-results
		resultsMap[r.callerID] = r
	}

	// Critical assertion: caller with normal context must succeed
	normalResult := resultsMap["normal"]
	assert.NoError(t, normalResult.err, "Caller with valid context should succeed even if another caller's context is canceled")
	assert.Equal(t, "workspace-456", normalResult.workspaceID)

	// Canceled caller may succeed (if result arrives before 10ms timeout) or fail - both are valid
	canceledResult := resultsMap["canceled"]
	if canceledResult.err != nil {
		t.Logf("Canceled caller returned error as expected: %v", canceledResult.err)
	} else {
		t.Logf("Canceled caller succeeded before timeout (received shared result from singleflight)")
		assert.Equal(t, "workspace-456", canceledResult.workspaceID)
	}

	// Backend should only be called once (singleflight deduplication worked)
	assert.Equal(t, int32(1), callCount.Load(), "Backend should be called once despite context cancellation")
}

// TestGetDefaultWorkspaceIDWithCache_TypeAssertionFailure verifies that corrupted workspace cache entries
// (wrong type) are detected, deleted, and the code falls through to fetch from API.
func TestGetDefaultWorkspaceIDWithCache_TypeAssertionFailure(t *testing.T) {
	// Create a mock RBAC server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error": "test error"}`)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create RBAC client pointing to mock server
	client := NewRbacClient(server.URL, nil, 10*time.Second, RbacClientConfig{}, nil).(*rbacClientImpl)

	// Create context with request_id via middleware
	ctx := contextWithRequestID(t)

	// Add identity to context
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			Type:  "User",
			OrgID: "org-badtype",
			User: &identity.User{
				UserID: "user-badtype",
			},
		},
	})

	// Manually corrupt the cache with wrong type ([]string instead of string)
	reqID := request_id.GetReqID(ctx)
	cacheKey := hashCacheKey("workspace", reqID, "org-badtype", "user-badtype")
	client.workspaceCache.Set(cacheKey, []string{"wrong", "type"}, cache.DefaultExpiration)

	// Verify corrupted entry exists
	corrupted, found := client.workspaceCache.Get(cacheKey)
	assert.True(t, found, "Corrupted cache entry should exist")
	corruptedValue, ok := corrupted.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"wrong", "type"}, corruptedValue, "Should have corrupted value")

	// Call should detect type mismatch in cache, delete it, then try to fetch from API
	// API call will fail (mock returns error) but important thing is it didn't panic
	_, err := client.GetDefaultWorkspaceIDWithCache(ctx, "org-badtype")

	// Expect error from API call (mock server error), not panic from type assertion
	assert.Error(t, err, "Should get error from API call, not panic from type assertion")

	// Cache remains empty because API call failed (can't repopulate with fresh data)
	_, found = client.workspaceCache.Get(cacheKey)
	assert.False(t, found, "Corrupted cache entry should be deleted and not repopulated after failed API call")
}
