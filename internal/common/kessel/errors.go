// Package kessel provides Kessel inventory client integration for workspace-based authorization.
//
// Coded in collaboration with AI
package kessel

import "errors"

// Error types for distinguishing between different failure modes
var (
	// ErrIdentityValidation indicates the identity header is malformed or missing required fields
	// Should result in HTTP 400 Bad Request
	ErrIdentityValidation = errors.New("identity validation failed")

	// ErrServiceUnavailable indicates a downstream service (Kessel, RBAC) is unavailable
	// Should result in HTTP 503 Service Unavailable
	ErrServiceUnavailable = errors.New("authorization service unavailable")
)

// IsIdentityValidationError checks if an error is an identity validation error
func IsIdentityValidationError(err error) bool {
	return errors.Is(err, ErrIdentityValidation)
}

// IsServiceUnavailableError checks if an error is a service unavailable error
func IsServiceUnavailableError(err error) bool {
	return errors.Is(err, ErrServiceUnavailable)
}
