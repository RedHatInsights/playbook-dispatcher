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

const shutdownTimeout = 20 * time.Second

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
	defer utils.CloseLogger()
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

	log.Infow("Playbook dispatcher started", "version", cfg.GetString("build.commit"))

	// stop on signal or error, whatever comes first
	select {
	case signal := <-signals:
		log.Infow("Shutting down", "signal", signal)
		return nil
	case error := <-errors:
		log.Errorw("Shutting down", "error", error)
		return error
	}
}

func shutdown(server *echo.Echo, log *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer func() {
		if err := log.Sync(); err != nil {
			log.Error(err)
		}
	}()

	defer log.Info("Shutdown complete")

	if err := utils.WgWaitFor(wg, shutdownTimeout); err != nil {
		log.Warn(err)
	}

	ctx, cancel := context.WithTimeout(utils.SetLog(context.Background(), log), shutdownTimeout)
	defer cancel()

	utils.StopServer(ctx, server)
}
