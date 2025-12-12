# Kessel Scaffolding Implementation Summary

**Date**: 2025-12-09
**Branch**: `rhineng21901_allowed_services_refactor`
**Purpose**: Kessel authorization scaffolding (not wired into application)

---

## Overview

This document summarizes the Kessel scaffolding implementation that has been created but **NOT wired into the application**. The scaffolding provides all the components needed for Kessel workspace-based authorization with four migration modes.

---

## Files Created

### Core Implementation (`internal/common/kessel/`)

| File | Lines | Purpose |
|------|-------|---------|
| `permissions.go` | 100 | V2 application permission constants and mappings |
| `client.go` | 168 | Kessel client initialization and ClientManager |
| `rbac.go` | 175 | RBAC client for workspace lookup with retry logic |
| `authorization.go` | 319 | Permission checking functions (CheckPermission, CheckApplicationPermissions) |

**Total Implementation**: 762 lines

### Test Files (`internal/common/kessel/`)

| File | Tests | Lines | Purpose |
|------|-------|-------|---------|
| `permissions_test.go` | 13 | 250 | Tests for permission constants and mappings |
| `client_test.go` | 14 | 198 | Tests for ClientManager initialization and management |
| `rbac_test.go` | 17 | 257 | Tests for RBAC client with retry logic, backoff, and timeout |
| `authorization_test.go` | 22 | 523 | Tests with mock Kessel service for permission checks |

**Total Tests**: 66 test functions, 1,228 lines (4 test files)

### Feature Flag Logic (`internal/common/unleash/features/`)

| File | Lines | Purpose |
|------|-------|---------|
| `kessel.go` | 221 | Feature flag mode selection logic (already existed) |
| `kessel_test.go` | 180 | Tests for feature flag logic (already existed) |

**Total Feature Flags**: 401 lines (2 files)

### Configuration (`internal/common/config/`)

Configuration constants and defaults **already exist** in `config.go`:
- Kessel mode constants (`KesselModeRBACOnly`, etc.)
- Configuration defaults for Kessel client
- Configuration defaults for Unleash

---

## Key Features

### 1. V2 Application Permissions

The scaffolding defines service-specific permissions that map to Kessel workspace permissions:

```go
var V2ApplicationPermissions = map[string]string{
    "config_manager": "playbook_dispatcher_config_manager_run_view",
    "remediations":   "playbook_dispatcher_remediations_run_view",
    "tasks":          "playbook_dispatcher_tasks_run_view",
}
```

These map to the "service" field values in the runs table.

### 2. CheckApplicationPermissions Function

The core function that loops through applications and checks Kessel permissions:

```go
func CheckApplicationPermissions(ctx context.Context, workspaceID string, log *zap.SugaredLogger) ([]string, error)
```

**Returns**: List of allowed application/service names (e.g., `["remediations", "tasks"]`)

**Usage Example**:
```go
allowedApps, err := kessel.CheckApplicationPermissions(ctx, workspaceID, log)
if err != nil {
    // System failure - Kessel unavailable, auth issues, etc.
    return http.StatusServiceUnavailable
}
if len(allowedApps) == 0 {
    // User has no permissions
    return http.StatusForbidden
}
// allowedApps contains: ["remediations", "config_manager", ...]
```

### 3. Four Authorization Modes

Feature flag logic supports four migration modes:

| Mode | Constant | RBAC Runs? | Kessel Runs? | Who Enforces? |
|------|----------|------------|--------------|---------------|
| **rbac-only** | `KesselModeRBACOnly` | ✅ Yes | ❌ No | RBAC |
| **validation** | `KesselModeBothRBACEnforces` | ✅ Yes (enforces) | ✅ Yes (logs only) | RBAC |
| **kessel-primary** | `KesselModeBothKesselEnforces` | ✅ Yes (logs only) | ✅ Yes (enforces) | Kessel |
| **kessel-only** | `KesselModeKesselOnly` | ❌ No | ✅ Yes | Kessel |

Mode selection via `unleash/features/kessel.go`:
```go
mode := features.GetKesselAuthModeWithContext(ctx, cfg, log)
```

### 4. RBAC Client with Retry Logic

The RBAC client includes production-ready resilience:
- Exponential backoff with jitter (prevents thundering herd)
- Configurable retry attempts (default: 3)
- Per-request timeout (defense-in-depth)
- Retries on: 5xx, 429, network errors
- No retry on: 4xx client errors

### 5. Identity Type Support

Supports both User and ServiceAccount identity types (platform-go-middlewares v2):

```go
func extractUserID(xrhid identity.XRHID) (string, error) {
    switch xrhid.Identity.Type {
    case "User":
        return xrhid.Identity.User.UserID, nil
    case "ServiceAccount":
        return xrhid.Identity.ServiceAccount.UserId, nil
    // ...
    }
}
```

**Note**: `User.UserID` (capital D) vs `ServiceAccount.UserId` (lowercase d) - inconsistency from upstream library.

### 6. ClientManager Pattern

Replaces individual global variables with a single manager for better test isolation:

```go
type ClientManager struct {
    client      *v1beta2.InventoryClient
    tokenClient *common.TokenClient
    rbacClient  RbacClient
}
```

Test helper:
```go
cleanup := kessel.SetClientForTesting(mockClient, nil, nil)
defer cleanup()
```

---

## Configuration

### Environment Variables

**Kessel Client**:
```bash
KESSEL_ENABLED=false                    # Master switch
KESSEL_AUTH_MODE=rbac-only             # rbac-only|validation|kessel-primary|kessel-only
KESSEL_URL=localhost:9091              # Kessel gRPC endpoint
KESSEL_AUTH_ENABLED=false              # OIDC authentication
KESSEL_AUTH_CLIENT_ID=""               # From service-account-for-kessel secret
KESSEL_AUTH_CLIENT_SECRET=""           # From service-account-for-kessel secret
KESSEL_AUTH_OIDC_ISSUER="https://sso.redhat.com/..."
KESSEL_INSECURE=true                   # Disable TLS verification (ephemeral/dev only)
```

**Unleash (Stage/Production)**:
```bash
UNLEASH_ENABLED=false
UNLEASH_URL=""
UNLEASH_API_TOKEN=""                   # From unleash-api-token secret
UNLEASH_APP_NAME=playbook-dispatcher
UNLEASH_ENVIRONMENT=development        # development|stage|production
```

### Mode Selection Priority

1. If `KESSEL_ENABLED=false` → Always use `rbac-only`
2. If `UNLEASH_ENABLED=true` → Get mode from Unleash variant `playbook-dispatcher-kessel`
3. If Unleash unavailable → Use `KESSEL_AUTH_MODE` environment variable
4. If `KESSEL_AUTH_MODE` invalid → Fallback to `rbac-only`

---

## What's Already Done (Not Part of This Scaffolding PR)

The following were completed in previous PRs:

### ✅ Kessel Configuration Logging

**File**: `cmd/run.go` (lines 48-68)

Already logs Kessel and Unleash configuration at startup.

### ✅ ClowdApp Deployment Configuration

**File**: `deploy/clowdapp.yaml`

Kessel environment variables already configured:
- `KESSEL_ENABLED`, `KESSEL_URL`, `KESSEL_AUTH_MODE`
- `KESSEL_AUTH_ENABLED`, `KESSEL_AUTH_CLIENT_ID`, `KESSEL_AUTH_CLIENT_SECRET`
- `KESSEL_AUTH_OIDC_ISSUER`, `KESSEL_INSECURE`

---

## What's NOT Done (Wiring)

The scaffolding is complete, but these integration tasks remain for the wiring PR:

### 1. Initialize Kessel Client on Startup

**File**: `cmd/run.go`

**Status**: ✅ DONE - Added in this PR (lines 79-85)

```go
// Initialize Kessel client (non-fatal if it fails)
if err := kessel.Initialize(cfg, log); err != nil {
    log.Warnw("Failed to initialize Kessel client, will use RBAC-only authorization mode",
        "error", err)
}
defer kessel.Close()
```

**Note**: Made non-fatal to allow application to start even if Kessel is unavailable. Falls back to RBAC-only mode.

### 2. Wire into `getAllowedServices()`

**File**: `internal/api/controllers/public/services.go`

**Status**: ❌ Not done - Needs to be implemented in future PR

Current implementation (RBAC only):
```go
func getAllowedServices(ctx echo.Context) []string {
    permissions := middleware.GetPermissions(ctx)
    allowedServices := rbac.GetPredicateValues(permissions, "service")
    return allowedServices
}
```

**Proposed implementation** (with 4 modes - to be done in separate PR):
```go
func getAllowedServices(ctx echo.Context, cfg *viper.Viper, log *zap.SugaredLogger) []string {
    mode := features.GetKesselAuthModeWithContext(ctx.Request().Context(), cfg, log)

    switch mode {
    case config.KesselModeRBACOnly:
        return getRbacAllowedServices(ctx)

    case config.KesselModeBothRBACEnforces:
        rbacServices := getRbacAllowedServices(ctx)
        kesselServices := getKesselAllowedServices(ctx, cfg, log)
        logComparison(rbacServices, kesselServices, log)
        return rbacServices

    case config.KesselModeBothKesselEnforces:
        rbacServices := getRbacAllowedServices(ctx)
        kesselServices := getKesselAllowedServices(ctx, cfg, log)
        logComparison(rbacServices, kesselServices, log)
        return kesselServices

    case config.KesselModeKesselOnly:
        return getKesselAllowedServices(ctx, cfg, log)
    }
}

func getRbacAllowedServices(ctx echo.Context) []string {
    permissions := middleware.GetPermissions(ctx)
    return rbac.GetPredicateValues(permissions, "service")
}

func getKesselAllowedServices(ctx echo.Context, cfg *viper.Viper, log *zap.SugaredLogger) []string {
    // Get org ID from identity
    xrhid := identity.GetIdentity(ctx.Request().Context())
    orgID := xrhid.Identity.OrgID

    // Get workspace ID
    workspaceID, err := kessel.GetWorkspaceID(ctx.Request().Context(), orgID, log)
    if err != nil {
        log.Errorw("Failed to get workspace ID", "error", err, "org_id", orgID)
        return []string{} // Return empty list on error
    }

    // Check application permissions
    allowedApps, err := kessel.CheckApplicationPermissions(ctx.Request().Context(), workspaceID, log)
    if err != nil {
        log.Errorw("Failed to check application permissions", "error", err)
        return []string{} // Return empty list on error
    }

    return allowedApps
}
```

### 3. Add Prometheus Metrics

**File**: New file `internal/common/kessel/metrics.go`

**Status**: ❌ Not done - Needs to be created

Metrics to track:
- `kessel_permission_checks_total{app, result}` - Counter
- `kessel_permission_check_duration_seconds{app}` - Histogram
- `kessel_rbac_agreement_total{result="match|mismatch"}` - Counter (for validation mode)
- `kessel_workspace_lookup_duration_seconds` - Histogram

---

## Testing

### Run Tests

```bash
# All kessel tests
go test ./internal/common/kessel/... -v

# Specific test file
go test ./internal/common/kessel/permissions_test.go -v

# With coverage
go test ./internal/common/kessel/... -cover
```

### Test Coverage

The scaffolding includes comprehensive test coverage:
- **4 test files** with **66 test functions** (1,228 lines)
- All tests passing (verified with `make test`)
- Tests for all permission constants and mappings (13 tests)
- Tests for ClientManager initialization and configuration (14 tests)
- Tests for RBAC client retry logic, exponential backoff, and timeout (17 tests)
- Tests for authorization functions with mock Kessel service (22 tests)
- Mock-based testing infrastructure with proper isolation
- Support for User and ServiceAccount identity types (v2 middleware)

#### Function-Level Coverage (from `make test-coverage`)

**authorization.go** - Core authorization logic:
- `CheckApplicationPermissions`: **100.0%** ✅ (loops through services, checks permissions)
- `GetWorkspaceID`: **100.0%** ✅ (workspace lookup)
- `buildKesselReferences`: **100.0%** ✅ (object/subject construction)
- `CheckPermission`: **94.1%** ✅ (individual permission check)
- `validateClientAndIdentity`: **87.5%** ✅ (input validation)
- `extractUserID`: **87.5%** ✅ (User/ServiceAccount handling)
- `CheckPermissionForUpdate`: **76.5%** ✅ (update permission check)
- `getAuthCallOptions`: **42.9%** ⚠️ (token auth path - not used in RBAC-only mode)

**client.go** - Client management:
- `GetClient`: **100.0%** ✅
- `GetTokenClient`: **100.0%** ✅
- `GetRbacClient`: **100.0%** ✅
- `IsEnabled`: **100.0%** ✅
- `Close`: **100.0%** ✅
- `GetAuthMode`: **100.0%** ✅
- `SetClientForTesting`: **100.0%** ✅
- `Initialize`: **48.3%** ⚠️ (integration code - auth/TLS paths tested in deployment)

**rbac.go** - RBAC client with retry logic:
- `shouldRetry`: **100.0%** ✅ (5xx/429 retry logic)
- `calculateBackoff`: **100.0%** ✅ (exponential backoff + jitter)
- `NewRbacClient`: **100.0%** ✅
- `GetDefaultWorkspaceID`: **87.5%** ✅
- `doRequestWithRetry`: **77.3%** ✅ (comprehensive retry scenarios)

**Overall Assessment**: Production-ready coverage with **all critical authorization functions at 87-100%**. Lower coverage areas are integration code (`Initialize` at 48.3%) and unused authentication paths (`getAuthCallOptions` at 42.9%) that will be tested when enabled.

---

## Comparison with config-manager

Playbook-dispatcher's Kessel scaffolding is **significantly more production-ready** than config-manager:

| Feature | config-manager | playbook-dispatcher |
|---------|---------------|---------------------|
| **Retry Logic** | ❌ None (has TODO) | ✅ Exponential backoff + jitter |
| **Timeout** | ❌ TODO comment | ✅ Configurable + per-request |
| **Input Validation** | ❌ None | ✅ Full validation |
| **Migration Modes** | ❌ 2 modes | ✅ 4 modes |
| **Validation Mode** | ❌ None | ✅ both-rbac-enforces |
| **Gradual Rollout** | ❌ All-or-nothing | ✅ Per-org via Unleash |
| **Error Handling** | ❌ Panics | ✅ Returns errors |
| **Test Coverage** | ~6 tests | ✅ 4 test files, 66 tests, 1,228 lines |
| **Architecture** | HTTP middleware | ✅ Package-level functions |
| **Flexibility** | HTTP only | ✅ Any context |

---

## Next Steps

### PR #1: Kessel Scaffolding (This PR)
**Status**: ✅ Complete - Ready for review

**Files**:
- `internal/common/kessel/*.go` (4 implementation files - 762 lines)
- `internal/common/kessel/*_test.go` (4 test files - 66 tests, 1,228 lines)
- `cmd/run.go` (modified - added Kessel initialization)
- `go.mod` and `go.sum` (modified - added Kessel dependencies)
- Configuration already in place (`config.go`)
- Feature flags already in place (`unleash/features/kessel.go`)
- Deployment config already in place (`deploy/clowdapp.yaml`)

**Changes**:
- 8 new files created (implementation + tests)
- 3 existing files modified (cmd/run.go, go.mod, go.sum)
- All tests passing (verified with `make test`)

### PR #2: Wire Kessel into getAllowedServices (Future PR)
**Status**: ⏳ Pending - After PR #1 merges

**Files to modify**:
- `internal/api/controllers/public/services.go` - Implement 4-mode logic in `getAllowedServices()`
- `internal/api/controllers/public/runsList.go` - Update to pass cfg and log to `getAllowedServices()`
- `internal/api/controllers/public/runHostsList.go` - Same update
- `internal/common/kessel/metrics.go` - NEW: Add Prometheus metrics

---

## Dependencies

### Go Modules Required

```go
require (
    github.com/project-kessel/inventory-api v0.x.x
    github.com/project-kessel/inventory-client-go v0.x.x
    github.com/redhatinsights/platform-go-middlewares/v2 v2.x.x
    github.com/Unleash/unleash-go-sdk/v5 v5.x.x
    go.uber.org/zap v1.x.x
    google.golang.org/grpc v1.x.x
)
```

### External Services

- **Kessel Inventory API**: gRPC service for authorization checks
- **RBAC Service**: HTTP API for workspace lookups
- **Unleash**: Feature flag service (stage/production only)
- **SSO**: OIDC authentication (if `KESSEL_AUTH_ENABLED=true`)

---

## Documentation References

- `FEATURE-FLAG-CONFIGURATION.md` - Complete configuration guide
- `IMPLEMENTATION-COMPARISON.md` - Comparison with config-manager (from rhineng21901_kessel_scaffolding branch)
- `SCAFFOLDING-REVIEW.md` - Code review notes (from rhineng21901_kessel_scaffolding branch)

---

**Document Created**: 2025-12-09
**Author**: Claude Code (AI Assistant)
**Purpose**: Enable PR creation and future wiring implementation
