// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"playbook-dispatcher/internal/common/utils"

	"github.com/patrickmn/go-cache"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
	"go.uber.org/zap"
)

// KesselClientWithCache wraps the Kessel inventory client with permission caching
type KesselClientWithCache struct {
	client          *v1beta2.InventoryClient
	permissionCache *cache.Cache
}

// NewKesselClientWithCache creates a new Kessel client wrapper with caching
func NewKesselClientWithCache(client *v1beta2.InventoryClient) *KesselClientWithCache {
	return &KesselClientWithCache{
		client:          client,
		permissionCache: cache.New(1*time.Minute, 30*time.Second),
	}
}

func getUserIDFromContext(ctx context.Context) string {
	xrhid := identity.GetIdentity(ctx)

	// Extract user ID based on identity type
	identityType := xrhid.Identity.Type
	switch identityType {
	case "User":
		if xrhid.Identity.User != nil {
			return xrhid.Identity.User.UserID
		}
	case "ServiceAccount":
		if xrhid.Identity.ServiceAccount != nil {
			return xrhid.Identity.ServiceAccount.UserId
		}
	}

	return ""
}

// hashCacheKey creates a SHA256 hash of the combined cache key components
// Uses colon as delimiter to prevent ambiguity (e.g., "a"+"bc" vs "ab"+"c")
func hashCacheKey(parts ...string) string {
	combined := ""
	for i, part := range parts {
		if i > 0 {
			combined += ":"
		}
		combined += part
	}
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

func (r *rbacClientImpl) GetDefaultWorkspaceIDWithCache(ctx context.Context, orgID string) (string, error) {
	reqID := request_id.GetReqID(ctx)
	userID := getUserIDFromContext(ctx)

	// Validate all cache key components (security: prevent cache leakage)
	if reqID == "" {
		return "", errors.New("request_id is required for caching")
	}
	if orgID == "" {
		return "", errors.New("org_id is required for caching")
	}
	if userID == "" || userID == "unknown" {
		return "", errors.New("valid user_id is required for caching")
	}

	// NOTE: reqID can be externally provided by the calling application and may be
	// the same across multiple endpoint requests, enabling cross-request caching
	cacheKey := hashCacheKey("workspace", reqID, orgID, userID)

	// Check cache
	if cached, found := r.workspaceCache.Get(cacheKey); found {
		log := utils.GetLogFromContextIfAvailable(ctx)
		workspaceID, ok := cached.(string)
		if !ok {
			// Type mismatch - delete bad entry and fall through to API
			if log != nil {
				log.Warnw("Workspace cache type mismatch, deleting entry",
					"request_id", reqID,
					"internal_request_id", utils.GetInternalRequestID(ctx),
					"org_id", orgID,
					"cache_key", cacheKey)
			}
			r.workspaceCache.Delete(cacheKey)
			// Fall through to fetch from API
		} else {
			if log != nil {
				log.Debugw("Workspace cache hit",
					"request_id", reqID,
					"internal_request_id", utils.GetInternalRequestID(ctx),
					"org_id", orgID,
					"cache_key", cacheKey)
			}
			return workspaceID, nil
		}
	}

	// Cache miss - fetch from API
	workspaceID, err := r.GetDefaultWorkspaceID(ctx, orgID)
	if err != nil {
		return "", err
	}

	// Store in cache
	r.workspaceCache.Set(cacheKey, workspaceID, cache.DefaultExpiration)

	return workspaceID, nil
}

// CheckPermissionWithCache performs a Kessel authorization check with caching
// Cache key is a hash of: request_id + org_id + user_id + workspace_id + permission
func (k *KesselClientWithCache) CheckPermissionWithCache(ctx context.Context, workspaceID string, permission string, log *zap.SugaredLogger) (bool, error) {
	reqID := request_id.GetReqID(ctx)
	xrhid := identity.GetIdentity(ctx)
	orgID := xrhid.Identity.OrgID
	userID := getUserIDFromContext(ctx)

	// Validate all cache key components (security: prevent cache leakage)
	if reqID == "" {
		return false, errors.New("request_id is required for caching")
	}
	if orgID == "" {
		return false, errors.New("org_id is required for caching")
	}
	if userID == "" || userID == "unknown" {
		return false, errors.New("valid user_id is required for caching")
	}
	if workspaceID == "" {
		return false, errors.New("workspace_id is required for caching")
	}
	if permission == "" {
		return false, errors.New("permission is required for caching")
	}

	// NOTE: reqID can be externally provided by the calling application and may be
	// the same across multiple endpoint requests, enabling cross-request caching
	cacheKey := hashCacheKey("permission", reqID, orgID, userID, workspaceID, permission)

	// Check cache
	if cached, found := k.permissionCache.Get(cacheKey); found {
		allowed, ok := cached.(bool)
		if !ok {
			// Type mismatch - delete bad entry and fall through to Kessel check
			if log != nil {
				log.Warnw("Permission cache type mismatch, deleting entry",
					"request_id", reqID,
					"internal_request_id", utils.GetInternalRequestID(ctx),
					"org_id", orgID,
					"permission", permission,
					"cache_key", cacheKey)
			}
			k.permissionCache.Delete(cacheKey)
			// Fall through to check permission via Kessel
		} else {
			if log != nil {
				log.Debugw("Permission cache hit",
					"request_id", reqID,
					"internal_request_id", utils.GetInternalRequestID(ctx),
					"org_id", orgID,
					"permission", permission,
					"cache_key", cacheKey)
			}
			return allowed, nil
		}
	}

	// Cache miss - check permission via Kessel
	allowed, err := CheckPermission(ctx, workspaceID, permission, log)
	if err != nil {
		return false, err
	}

	// Store in cache (cache both allowed and denied results)
	k.permissionCache.Set(cacheKey, allowed, cache.DefaultExpiration)

	return allowed, nil
}
