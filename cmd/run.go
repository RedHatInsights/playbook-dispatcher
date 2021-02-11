package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"playbook-dispatcher/internal/api"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils"
	responseConsumer "playbook-dispatcher/internal/response-consumer"
	"playbook-dispatcher/internal/validator"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type startModuleFn = func(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
)

func run(cmd *cobra.Command, args []string) error {
	modules, err := cmd.Flags().GetStringSlice("module")
	utils.DieOnError(err)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	errors := make(chan error, 1)

	log := utils.GetLoggerOrDie()
	cfg := config.Get()

	metricsServer := echo.New()
	metricsServer.HideBanner = true
	metricsServer.Debug = false

	readinessProbeHandler := &utils.ProbeHandler{}
	livenessProbeHandler := &utils.ProbeHandler{}

	metricsServer.GET("/ready", readinessProbeHandler.Check)
	metricsServer.GET("/live", livenessProbeHandler.Check)
	metricsServer.GET(cfg.GetString("metrics.path"), echo.WrapHandler(promhttp.Handler()))

	wg := sync.WaitGroup{}

	ctx, stop := context.WithCancel(utils.SetLog(context.Background(), log))
	defer shutdown(metricsServer, log, &wg)
	defer stop()

	for _, module := range modules {
		log.Infof("Starting module %s", module)

		var startModule startModuleFn

		switch module {
		case moduleApi:
			startModule = api.Start
		case moduleResponseConsumer:
			startModule = responseConsumer.Start
		case moduleValidator:
			startModule = validator.Start
		default:
			return fmt.Errorf("Unknown module %s", module)
		}

		startModule(ctx, cfg, errors, readinessProbeHandler, livenessProbeHandler, &wg)
	}

	log.Infof("Listening on service port %d", cfg.GetInt("metrics.port"))
	go func() {
		errors <- metricsServer.Start(fmt.Sprintf("0.0.0.0:%d", cfg.GetInt("metrics.port")))
	}()

	log.Infow("Playbook dispatcher started", "version", cfg.GetString("openshift.build.commit"))

	// stop on signal or error, whatever comes first
	select {
	case signal := <-signals:
		log.Infow("Shutting down", "signal", signal)
		return nil
	case error := <-errors:
		log.Fatalw("Shutting down", "error", error)
		return error
	}
}

func shutdown(server *echo.Echo, log *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer log.Sync()
	defer log.Info("Shutdown complete")

	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	utils.StopServer(ctx, server)
}
