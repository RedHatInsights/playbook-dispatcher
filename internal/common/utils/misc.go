package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

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

func WgWaitFor(wg *sync.WaitGroup, timeout time.Duration) error {
	channel := make(chan struct{})

	go func() {
		wg.Wait()
		close(channel)
	}()

	select {
	case <-channel:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Time out waiting for WaitGroup")
	}
}
