package middleware

import (
	"playbook-dispatcher/internal/common/utils"
	"time"

	"github.com/labstack/echo/v4"
	identityMiddleware "github.com/redhatinsights/platform-go-middlewares/v2/identity"
)

func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log := utils.GetLogFromEcho(c)

		start := time.Now()

		// Log request start to help identify retries/timeouts
		req := c.Request()
		log.Debugw("request started",
			"method", req.Method,
			"url", req.RequestURI,
			"remote_addr", req.RemoteAddr,
			"x_forwarded_for", req.Header.Get("X-Forwarded-For"),
			"user_agent", req.UserAgent())

		err := next(c)

		res := c.Response()

		statusCode := res.Status

		if httpError, ok := err.(*echo.HTTPError); ok {
			statusCode = httpError.Code
		}

		identity := identityMiddleware.GetIdentity(c.Request().Context())
		if identity.Identity.OrgID != "" {
			log = log.With(
				"account", identity.Identity.AccountNumber,
				"org_id", identity.Identity.OrgID,
			)
		}

		log.Infow(
			"request completed",
			"time", time.Since(start).String(),
			"status", statusCode,
			"method", req.Method,
			"url", req.RequestURI,
			"referer", req.Referer(),
			"user_agent", req.UserAgent(),
		)

		return err
	}
}
