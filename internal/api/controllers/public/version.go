package public

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Version string `json:"version"`
}

func (this *controllers) ApiVersion(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, &Response{Version: "1.0"})
}
