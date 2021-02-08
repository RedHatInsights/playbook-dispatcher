package controllers

import (
	"net/http"
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
		this.log.Warn(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// send all requests to cloud connector concurrently
	result := input.PMapRunCreated(func(runInput RunInput) *RunCreated {
		recipient, err := uuid.Parse(string(runInput.Recipient))
		if err != nil {
			this.log.Error(err) // TODO: probes
			return runCreateError(http.StatusBadRequest)
		}

		messageId, err := this.cloudConnectorClient.SendCloudConnectorRequest(
			ctx.Request().Context(),
			string(runInput.Account),
			recipient,
		)

		if err != nil {
			this.log.Error(err)
			return runCreateError(http.StatusInternalServerError)
		}

		messageIdUuid, err := uuid.Parse(*messageId)
		if err != nil {
			this.log.Error(err)
			return runCreateError(http.StatusInternalServerError)
		}

		entity, err := newRun(&runInput, messageIdUuid, dbModel.RunStatusRunning)
		if err != nil {
			this.log.Error(err)
			return runCreateError(http.StatusInternalServerError)
		}

		if dbResult := this.database.Create(entity); dbResult.Error != nil {
			this.log.Error(dbResult.Error)
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

func newRun(input *RunInput, correlationId uuid.UUID, status string) (*dbModel.Run, error) {
	run := &dbModel.Run{
		ID:            uuid.New(),
		Account:       string(input.Account),
		CorrelationID: correlationId,
		URL:           string(input.Url),
		Status:        status,
	}

	if input.Labels != nil {
		run.Labels = input.Labels.AdditionalProperties
	}

	if input.Timeout != nil {
		run.Timeout = int(*input.Timeout)
	} else {
		run.Timeout = cfg.GetInt("default.run.timeout")
	}

	if recipient, err := uuid.Parse(string(input.Recipient)); err == nil {
		run.Recipient = recipient
	} else {
		return nil, err
	}

	return run, nil
}

func runCreateError(code int) *RunCreated {
	return &RunCreated{
		Code: code,
		// TODO report error
	}
}
