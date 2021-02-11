package api

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/controllers"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/utils"
	"sync"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/spf13/viper"
)

const specFile = "/api/playbook-dispatcher/v1/openapi.json"
const requestIdHeader = "x-rh-insights-request-id"

func init() {
	openapi3.DefineStringFormat("uuid", `^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}$`)
	openapi3.DefineStringFormat("url", `^https?:\/\/.*$`)
}

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
) {
	instrumentation.Start()
	db, sql := db.Connect(ctx, cfg)

	ready.Register(sql.Ping)
	live.Register(sql.Ping)

	spec, err := controllers.GetSwagger()
	utils.DieOnError(err)

	server := echo.New()
	server.HideBanner = true
	server.Debug = false

	server.Use(
		echoPrometheus.MetricsMiddleware(),
		echoMiddleware.BodyLimit(cfg.GetString("http.max.body.size")),
		echo.WrapMiddleware(request_id.ConfiguredRequestID(requestIdHeader)),
		middleware.ContextLogger,
		echoMiddleware.Recover(),
	)

	server.GET(specFile, func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, spec)
	})

	var cloudConnectorClient connectors.CloudConnectorClient

	if cfg.GetString("cloud.connector.impl") == "impl" {
		cloudConnectorClient = connectors.NewConnectorClient(cfg)
	} else {
		cloudConnectorClient = connectors.NewConnectorClientMock()
		log.Warn("Using mock CloudConnectorClient")
	}

	ctrl := controllers.CreateControllers(db, cloudConnectorClient)

	internal := server.Group("/internal/*")
	public := server.Group("/api/playbook-dispatcher/v1/*")

	internal.Use(oapiMiddleware.OapiRequestValidator(spec))
	controllers.RegisterHandlers(internal, ctrl)

	public.Use(echo.WrapMiddleware(identity.EnforceIdentity))
	public.Use(echo.WrapMiddleware(middleware.EnforceIdentityType))
	public.Use(middleware.Hack("filter", "labels"))
	public.Use(middleware.Hack("fields"))
	public.Use(oapiMiddleware.OapiRequestValidator(spec))
	controllers.RegisterHandlers(public, ctrl)

	wg.Add(1)
	go func() {
		errors <- server.Start(fmt.Sprintf("0.0.0.0:%d", cfg.GetInt("web.port")))
	}()

	go func() {
		defer wg.Done()
		defer log.Debug("API stopped")
		<-ctx.Done()

		log.Info("Shutting down API")
		utils.StopServer(ctx, server)
		if sqlConnection, err := db.DB(); err != nil {
			sqlConnection.Close()
		}
	}()
}
