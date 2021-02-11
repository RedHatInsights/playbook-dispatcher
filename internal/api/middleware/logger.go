package middleware

import (
	"playbook-dispatcher/internal/common/utils"
	"time"

	"github.com/labstack/echo/v4"
)

func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log := utils.GetLogFromEcho(c)

		start := time.Now()

		err := next(c)

		req := c.Request()
		res := c.Response()

		log.Infow(
			"request completed",
			"time", time.Since(start).String(),
			"status", res.Status,
			"method", req.Method,
			"url", req.RequestURI,
			"referer", req.Referer(),
			"user_agent", req.UserAgent(),
		)

		return err
	}
}
