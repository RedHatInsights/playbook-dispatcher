package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/db"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

//go:generate fungen -types RunInput,*RunCreated -methods PMap -package private -filename utils.gen.go
func (this *controllers) ApiInternalRunsCreate(ctx echo.Context) error {
	var input RunInputList

	db.SetLog(this.database, utils.GetLogFromEcho(ctx))
	defer db.ClearLog(this.database)

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}
	rl := this.rateLimiter
	// send all requests to cloud connector concurrently
	result := input.PMapRunCreated(func(runInput RunInput) *RunCreated {
		recipient, err := uuid.Parse(string(runInput.Recipient))
		if err != nil {
			instrumentation.InvalidRecipientId(ctx, string(runInput.Recipient), err)
			return runCreateError(http.StatusBadRequest)
		}

		correlationId := uuid.New()

		if this.config.GetBool("demo.mode") {
			correlationId = uuid.UUID{}
		}

		context := utils.WithCorrelationId(ctx.Request().Context(), correlationId.String())
		context = utils.WithAccount(context, string(runInput.Account))

		// take from the rate limit pool
		rl.Take()

		messageId, notFound, err := this.cloudConnectorClient.SendCloudConnectorRequest(
			context,
			string(runInput.Account),
			recipient,
			correlationId,
			string(runInput.Url),
		)

		if err != nil {
			instrumentation.CloudConnectorRequestError(context, err, recipient)
			return runCreateError(http.StatusInternalServerError)
		} else if notFound {
			instrumentation.CloudConnectorNoConnection(context, recipient)
			return runCreateError(http.StatusNotFound)
		} else {
			instrumentation.CloudConnectorOK(context, recipient, messageId)
		}

		entity := newRun(&runInput, correlationId, dbModel.RunStatusRunning, recipient, this.config)
		entity.Service = middleware.GetPSKPrincipal(context)

		if dbResult := this.database.Create(&entity); dbResult.Error != nil {
			instrumentation.PlaybookRunCreateError(context, dbResult.Error, &entity)
			return runCreateError(http.StatusInternalServerError)
		}

		if runInput.Hosts != nil {
			newHosts := make([]dbModel.RunHost, len(*runInput.Hosts))

			for i, inputHost := range *runInput.Hosts {
				newHosts[i] = dbModel.RunHost{
					ID:          uuid.New(),
					RunID:       entity.ID,
					InventoryID: nil,
					Host:        inputHost.AnsibleHost,
					Status:      dbModel.RunStatusRunning,
				}

				if inputHost.InventoryId != nil {
					inventoryID := uuid.MustParse(string(*inputHost.InventoryId))
					newHosts[i].InventoryID = &inventoryID
				}
			}

			if dbResult := this.database.Create(newHosts); dbResult.Error != nil {
				instrumentation.PlaybookRunHostCreateError(context, dbResult.Error, newHosts)
				return runCreateError(http.StatusInternalServerError)
			}
		}

		runId := public.RunId(entity.ID.String())
		instrumentation.RunCreated(context, recipient, entity.ID, string(runInput.Url), entity.Service)

		return &RunCreated{
			Code: http.StatusCreated,
			Id:   &runId,
		}
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}

func newRun(input *RunInput, correlationId uuid.UUID, status string, recipient uuid.UUID, cfg *viper.Viper) dbModel.Run {
	run := dbModel.Run{
		ID:            uuid.New(),
		Account:       string(input.Account),
		CorrelationID: correlationId,
		URL:           string(input.Url),
		Status:        status,
		Recipient:     recipient,
	}

	if input.Labels != nil {
		run.Labels = input.Labels.AdditionalProperties
	}

	if input.Timeout != nil {
		run.Timeout = int(*input.Timeout)
	} else {
		run.Timeout = cfg.GetInt("default.run.timeout")
	}

	return run
}

func runCreateError(code int) *RunCreated {
	return &RunCreated{
		Code: code,
		// TODO report error
	}
}
