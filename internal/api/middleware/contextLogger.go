package middleware

import (
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v5"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
)

func ContextLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		reqId := request_id.GetReqID(ctx)
		logger := utils.LogWithRequestId(utils.GetLoggerOrDie(), reqId)

		// Add internal request ID to logger context
		if internalReqId := utils.GetInternalRequestID(ctx); internalReqId != "" {
			logger = logger.With("internal_request_id", internalReqId)
		}

		req := c.Request()
		c.SetRequest(req.WithContext(utils.SetLog(req.Context(), logger)))
		return next(c)
	}
}
