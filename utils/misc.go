package utils

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func DieOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func StopServer(server *echo.Echo, ctx context.Context, log *zap.SugaredLogger) {
	if e := server.Shutdown(ctx); e != nil {
		log.Fatal(e)
	}
}
