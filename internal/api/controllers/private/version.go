package private

import (
	"net/http"
	"playbook-dispatcher/internal/common/config"

	"github.com/labstack/echo/v4"
)

func (this *controllers) ApiInternalVersion(ctx echo.Context) error {
	cfg := config.Get()

	return ctx.JSON(http.StatusOK, Version(cfg.GetString("build.commit")))
}
