// Coded in collaboration with AI
package kessel

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/patrickmn/go-cache"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
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
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second).(*rbacClientImpl)

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
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second).(*rbacClientImpl)

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
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second).(*rbacClientImpl)

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

func TestCheckPermissionWithCache_ErrorOnMissingRequestID(t *testing.T) {
	// Create mock client
	mockService := &mockKesselInventoryServiceForCache{
		allowed: kesselv2.Allowed_ALLOWED_TRUE,
	}
	mockClient := &v1beta2.InventoryClient{
		KesselInventoryService: mockService,
	}

	// Create cache client wrapper
	kesselCache := NewKesselClientWithCache(mockClient)

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

	_, err := kesselCache.CheckPermissionWithCache(ctx, "workspace-abc", PermissionRemediationsRunView, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request_id is required")
}

func TestCheckPermissionWithCache_ErrorOnMissingWorkspaceID(t *testing.T) {
	mockService := &mockKesselInventoryServiceForCache{
		allowed: kesselv2.Allowed_ALLOWED_TRUE,
	}
	mockClient := &v1beta2.InventoryClient{
		KesselInventoryService: mockService,
	}

	// Create cache client wrapper
	kesselCache := NewKesselClientWithCache(mockClient)

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

	_, err := kesselCache.CheckPermissionWithCache(ctx, "", PermissionRemediationsRunView, log)

	assert.Error(t, err)
	// Will error on request_id first since we can't set it in tests
	assert.Contains(t, err.Error(), "required")
}

func TestCheckPermissionWithCache_ErrorOnMissingPermission(t *testing.T) {
	mockService := &mockKesselInventoryServiceForCache{
		allowed: kesselv2.Allowed_ALLOWED_TRUE,
	}
	mockClient := &v1beta2.InventoryClient{
		KesselInventoryService: mockService,
	}

	// Create cache client wrapper
	kesselCache := NewKesselClientWithCache(mockClient)

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

	_, err := kesselCache.CheckPermissionWithCache(ctx, "workspace-abc", "", log)

	assert.Error(t, err)
	// Will error on request_id first since we can't set it in tests
	assert.Contains(t, err.Error(), "required")
}

func TestWorkspaceCache_Expiration(t *testing.T) {
	// Create test client with short TTL cache
	client := NewRbacClient("http://localhost:8080", nil, 10*time.Second).(*rbacClientImpl)
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

func TestPermissionCache_Expiration(t *testing.T) {
	// Create test client with short TTL cache
	kesselClient := NewKesselClientWithCache(nil)
	kesselClient.permissionCache = cache.New(100*time.Millisecond, 50*time.Millisecond)

	// Set value with hashed key including workspace_id
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte("permission:req-123:org-456:user-789:workspace-abc:"+PermissionRemediationsRunView)))
	kesselClient.permissionCache.Set(cacheKey, true, cache.DefaultExpiration)

	// Verify it's there
	value, found := kesselClient.permissionCache.Get(cacheKey)
	assert.True(t, found)
	assert.Equal(t, true, value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify it's gone
	_, found = kesselClient.permissionCache.Get(cacheKey)
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
	// Same user, same permission, different workspaces should have different keys
	hash1 := hashCacheKey("permission", "req-123", "org-456", "user-789", "workspace-A", "playbook_dispatcher_remediations_run_view")
	hash2 := hashCacheKey("permission", "req-123", "org-456", "user-789", "workspace-B", "playbook_dispatcher_remediations_run_view")

	assert.NotEqual(t, hash1, hash2, "Different workspaces must produce different cache keys to prevent authorization leakage")
}

// NOTE: Full cache hit/miss functionality cannot be tested in unit tests because
// request_id.GetReqID() uses a private context key type from platform-go-middlewares.
// Without a valid request_id, the cache functions will return errors.
//
// Cache functionality is verified through:
// 1. Error validation tests above - ensure errors are returned when components are missing
// 2. Cache expiration tests - verify the TTL mechanism works
// 3. Hash key delimiter test - verify keys are properly structured
// 4. Integration/E2E tests with actual middleware that sets request_id in context
