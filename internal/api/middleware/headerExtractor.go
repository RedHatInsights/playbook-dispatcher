package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
)

/*
 * Extracts the given headers and stores them in context under the same key(s).
 */
func ExtractHeaders(headers ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			for _, header := range headers {
				ctx = context.WithValue(ctx, header, req.Header.Get(header)) //nolint:staticcheck
			}

			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func GetExtractedHeader(ctx context.Context, key string) string {
	if value, ok := ctx.Value(key).(string); !ok {
		return ""
	} else {
		return value
	}
}
