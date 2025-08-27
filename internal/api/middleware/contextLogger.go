package middleware

import (
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
)

func ContextLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqId := request_id.GetReqID(c.Request().Context())
		logger := utils.LogWithRequestId(utils.GetLoggerOrDie(), reqId)

		req := c.Request()
		c.SetRequest(req.WithContext(utils.SetLog(req.Context(), logger)))
		return next(c)
	}
}
