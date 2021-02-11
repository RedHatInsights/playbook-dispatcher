package utils

import (
	"context"

	"github.com/labstack/echo/v4"
)

func DieOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func StopServer(ctx context.Context, server *echo.Echo) {
	if e := server.Shutdown(ctx); e != nil {
		GetLogFromContext(ctx).Fatal(e)
	}
}

func IndexStrings(values ...string) map[string]string {
	result := make(map[string]string, len(values))

	for _, value := range values {
		result[value] = value
	}

	return result
}

func MapStrings(values []string, fn func(string) string) []string {
	result := make([]string, len(values))

	for i, value := range values {
		result[i] = fn(value)
	}

	return result
}
