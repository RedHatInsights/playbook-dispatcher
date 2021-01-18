package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

	log := utils.GetLoggerOrDie()
	cfg := config.Get()

	metricsServer := echo.New()
	metricsServer.HideBanner = true
	metricsServer.Debug = true

	// TODO
	probeHandler := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	metricsServer.GET("/live", probeHandler)
	metricsServer.GET("/ready", probeHandler)
	metricsServer.GET(cfg.GetString("metrics.path"), echo.WrapHandler(promhttp.Handler()))

	log.Infof("Listening on port %d", cfg.GetInt("metrics.port"))
	go func() {
		log.Fatal(metricsServer.Start(fmt.Sprintf("0.0.0.0:%d", cfg.GetInt("metrics.port"))))
	}()

	log.Infow("Playbook dispatcher started", "version", cfg.GetString("openshift.build.commit"))

	defer shutdown(metricsServer, log)

	<-signals
}

func shutdown(server *echo.Echo, log *zap.SugaredLogger) {
	defer log.Sync()
	defer log.Info("Shutdown complete")

	log.Info("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if e := server.Shutdown(ctx); e != nil {
		log.Fatal(e)
	}
}
