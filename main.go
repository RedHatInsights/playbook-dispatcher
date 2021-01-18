package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"playbook-dispatcher/api"
	"playbook-dispatcher/config"
	"playbook-dispatcher/utils"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	errors := make(chan error, 1)

	log := utils.GetLoggerOrDie()
	cfg := config.Get()

	metricsServer := echo.New()
	metricsServer.HideBanner = true
	metricsServer.Debug = false

	// TODO
	probeHandler := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	metricsServer.GET("/live", probeHandler)
	metricsServer.GET("/ready", probeHandler)
	metricsServer.GET(cfg.GetString("metrics.path"), echo.WrapHandler(promhttp.Handler()))

	stopApi := api.Start(cfg, log, errors)

	log.Infof("Listening on service port %d", cfg.GetInt("metrics.port"))
	go func() {
		err := metricsServer.Start(fmt.Sprintf("0.0.0.0:%d", cfg.GetInt("metrics.port")))
		log.Fatal(err)
		errors <- err
	}()

	log.Infow("Playbook dispatcher started", "version", cfg.GetString("openshift.build.commit"))

	defer shutdown(metricsServer, log, stopApi)

	// stop on signal or error, whatever comes first
	select {
	case signal := <-signals:
		log.Infow("Shutting down", "signal", signal)
	case error := <-errors:
		log.Infow("Shutting down", "error", error)
	}
}

func shutdown(server *echo.Echo, log *zap.SugaredLogger, stopApi func(context.Context)) {
	defer log.Sync()
	defer log.Info("Shutdown complete")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stopApi(ctx)
	utils.StopServer(server, ctx, log)
}
