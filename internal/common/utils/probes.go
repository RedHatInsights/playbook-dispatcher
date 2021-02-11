package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ProbeHandler struct {
	fns []func() error
}

func (this *ProbeHandler) Register(callback func() error) {
	this.fns = append(this.fns, callback)
}

func (this *ProbeHandler) Check(ctx echo.Context) error {
	for _, fn := range this.fns {
		if err := fn(); err != nil {
			GetLogFromEcho(ctx).Error(err)
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
	}

	return ctx.NoContent(http.StatusOK)
}
