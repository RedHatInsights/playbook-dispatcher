package api

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/controllers/private"
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/api/rbac"
	"playbook-dispatcher/internal/common/constants"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/utils"
	"sync"
	"time"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/spf13/viper"
)

const specFile = "/api/playbook-dispatcher/v1/openapi.json"
const apiShutdownTimeout = 10 * time.Second

func init() {
	openapi3.DefineStringFormat("uuid", `^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}$`)
	openapi3.DefineStringFormat("sat-id-uuid", `^[a-f0-9]{8}-[a-f0-9]{4}-[45][a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}$`)
	openapi3.DefineStringFormat("url", `^https?:\/\/.*$`)
}

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
	ready, live *utils.ProbeHandler,
	wg *sync.WaitGroup,
) {
	log := utils.GetLogFromContext(ctx)
	instrumentation.Start()
	db, sql := db.Connect(ctx, cfg)

	ready.Register(sql.Ping)
	live.Register(sql.Ping)

	publicSpec, err := public.GetSwagger()
	utils.DieOnError(err)

	privateSpec, err := private.GetSwaggerWithExternalRefs()
	utils.DieOnError(err)

	server := echo.New()
	server.HideBanner = true
	server.Debug = false

	server.Use(
		echoPrometheus.MetricsMiddleware(),
		echo.WrapMiddleware(request_id.ConfiguredRequestID(constants.HeaderRequestId)),
		middleware.ContextLogger,
		middleware.RequestLogger,
		echoMiddleware.Recover(),
		echoMiddleware.BodyLimit(cfg.GetString("http.max.body.size")),
	)

	server.GET(specFile, func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, publicSpec)
	})

	var cloudConnectorClient connectors.CloudConnectorClient

	if cfg.GetString("cloud.connector.impl") == "impl" {
		cloudConnectorClient = connectors.NewConnectorClient(cfg)
	} else {
		cloudConnectorClient = connectors.NewConnectorClientMock()
		log.Warn("Using mock CloudConnectorClient")
	}

	var translator tenantid.Translator
	if cfg.GetString("tenant.translator.impl") == "impl" {
		translator = tenantid.NewTranslator(
			fmt.Sprintf("%s://%s:%s", cfg.Get("tenant.translator.scheme"), cfg.Get("tenant.translator.host"), cfg.Get("tenant.translator.port")),
			tenantid.WithTimeout(cfg.GetDuration("tenant.translator.timeout")*time.Second),
			tenantid.WithMetrics(),
		)
	} else {
		translator = utils.NewDynamicMockTranslator()
		log.Warn("Using mock TenantIDTranslator")
	}

	authConfig := middleware.BuildPskAuthConfigFromEnv()
	log.Infow("Authentication required for internal API", "principals", utils.MapKeysString(authConfig))

	privateController := private.CreateController(db, cloudConnectorClient, cfg, translator)
	internal := server.Group("/internal")
	internal.Use(oapiMiddleware.OapiRequestValidator(privateSpec))
	// Authorization header not required for GET /internal/version
	internal.GET("/version", privateController.ApiInternalVersion)
	internal.Use(middleware.CheckPskAuth(authConfig))
	internal.POST("/dispatch", privateController.ApiInternalRunsCreate)
	internal.POST("/v2/recipients/status", privateController.ApiInternalV2RecipientsStatus)
	internal.POST("/v2/dispatch", privateController.ApiInternalV2RunsCreate)
	internal.POST("/v2/cancel", privateController.ApiInternalV2RunsCancel)

	publicController := public.CreateController(db, cloudConnectorClient)
	public := server.Group("/api/playbook-dispatcher")
	public.Use(echo.WrapMiddleware(identity.EnforceIdentity))
	public.Use(echo.WrapMiddleware(middleware.EnforceIdentityType))
	public.Use(middleware.CaptureQueryString())
	public.Use(middleware.Hack("filter", "labels"))
	public.Use(middleware.Hack("filter", "run"))
	public.Use(middleware.Hack("filter", "run", "labels"))
	public.Use(middleware.Hack("fields"))
	public.Use(oapiMiddleware.OapiRequestValidator(publicSpec))
	public.Use(middleware.ExtractHeaders(constants.HeaderIdentity))
	public.Use(middleware.EnforcePermissions(cfg, rbac.DispatcherPermission("run", "read")))

	public.GET("/v1/run_hosts", publicController.ApiRunHostsList)
	public.GET("/v1/runs", publicController.ApiRunsList)

	wg.Add(1)
	go func() {
		errors <- server.Start(fmt.Sprintf("0.0.0.0:%d", cfg.GetInt("web.port")))
	}()

	go func() {
		defer wg.Done()
		defer log.Debug("API stopped")
		<-ctx.Done()

		log.Info("Shutting down API")
		ctx, cancel := context.WithTimeout(utils.SetLog(context.Background(), log), apiShutdownTimeout)
		defer cancel()

		utils.StopServer(ctx, server)
		if sqlConnection, err := db.DB(); err != nil {
			sqlConnection.Close()
		}
	}()
}
