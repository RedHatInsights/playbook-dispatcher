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

func IndexStrings(values ...string) map[string]string {
	result := make(map[string]string, len(values))

	for _, value := range values {
		result[value] = value
	}

	return result
}
