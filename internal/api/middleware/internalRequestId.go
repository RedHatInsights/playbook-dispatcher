package middleware

import (
	"context"
	"playbook-dispatcher/internal/common/constants"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// InternalRequestId generates a unique server-side request ID for each request
// This is distinct from the client-provided request_id and is used for precise troubleshooting
func InternalRequestId(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		internalRequestId := uuid.New().String()

		// Store in context using typed key for type safety
		ctx := context.WithValue(c.Request().Context(), constants.InternalRequestIdContextKey, internalRequestId)
		c.SetRequest(c.Request().WithContext(ctx))

		// Add to response headers for client visibility and correlation
		c.Response().Header().Set(constants.HeaderInternalRequestId, internalRequestId)

		return next(c)
	}
}
