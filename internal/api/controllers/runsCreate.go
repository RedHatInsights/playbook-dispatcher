package controllers

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/common/config"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var cfg = config.Get()

//go:generate fungen -types RunInput,*RunCreated -methods PMap -package controllers -filename utils.gen.go
func (this *controllers) ApiInternalRunsCreate(ctx echo.Context) error {
	var input RunInputList

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// send all requests to cloud connector concurrently
	result := input.PMapRunCreated(func(runInput RunInput) *RunCreated {
		recipient, err := uuid.Parse(string(runInput.Recipient))
		if err != nil {
			instrumentation.InvalidRecipientId(ctx, string(runInput.Recipient), err)
			return runCreateError(http.StatusBadRequest)
		}

		correlationId := uuid.New()

		if cfg.GetBool("demo.mode") {
			correlationId = uuid.UUID{}
		}

		messageId, err := this.cloudConnectorClient.SendCloudConnectorRequest(
			ctx.Request().Context(),
			string(runInput.Account),
			recipient,
			correlationId,
			string(runInput.Url),
		)

		if err != nil {
			instrumentation.CloudConnectorRequestError(ctx, err, recipient)
			return runCreateError(http.StatusInternalServerError)
		} else {
			instrumentation.CloudConnectorOK(ctx, recipient, messageId)
		}

		entity := newRun(&runInput, correlationId, dbModel.RunStatusRunning, recipient)

		if dbResult := this.database.Create(&entity); dbResult.Error != nil {
			instrumentation.PlaybookRunCreateError(ctx, dbResult.Error, &entity)
			return runCreateError(http.StatusInternalServerError)
		}

		runId := RunId(entity.ID.String())
		return &RunCreated{
			Code: http.StatusCreated,
			Id:   &runId,
		}
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}

func newRun(input *RunInput, correlationId uuid.UUID, status string, recipient uuid.UUID) dbModel.Run {
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
