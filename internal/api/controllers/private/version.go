package private

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (this *controllers) ApiInternalVersion(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, Version("v2"))
}
