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

func (this *controllers) ApiInternalRunsCreate(ctx echo.Context) error {
	var input []RunInput

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		this.log.Warn(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	result := []*RunCreated{}

	for _, runInput := range input {
		// TODO: here we'll be making a call to cloud connector which gives us the correlation id
		correlationId := uuid.New()
		entity, err := newRun(&runInput, correlationId, dbModel.RunStatusRunning)
		if err != nil {
			result = append(result, &RunCreated{
				Code: http.StatusInternalServerError,
				// TODO report error
			})

			continue
		}

		if dbResult := this.database.Create(entity); dbResult.Error != nil {
			this.log.Error(dbResult.Error)
			result = append(result, &RunCreated{
				Code: http.StatusInternalServerError,
				// TODO report error
			})

			continue
		}

		runId := RunId(entity.ID.String())
		result = append(result, &RunCreated{
			Code: http.StatusCreated,
			Id:   &runId,
		})
	}

	return ctx.JSON(http.StatusMultiStatus, result)
}

func newRun(input *RunInput, correlationId uuid.UUID, status string) (*dbModel.Run, error) {
	run := &dbModel.Run{
		ID:            uuid.New(),
		Account:       string(input.Account),
		CorrelationID: correlationId,
		PlaybookURL:   string(input.Url),
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
