package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//go:generate fungen -types RunInput,*RunCreated -methods PMap -package private -filename utils.gen.go
func (this *controllers) ApiInternalRunsCreate(ctx echo.Context) error {
	var input RunInputList
	requestType := "ansible"

	db.SetLog(this.database, utils.GetLogFromEcho(ctx))
	defer db.ClearLog(this.database)

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// send all requests to cloud connector concurrently
	result := input.PMapRunCreated(func(runInputV1 RunInput) *RunCreated {
		runInput := RunInputV1GenericMap(runInputV1, this.config)

		recipient, err := uuid.Parse(runInput.Recipient)
		if err != nil {
			instrumentation.InvalidRecipientId(ctx, runInput.Recipient, err, requestType)
			return runCreateError(http.StatusBadRequest)
		}

		context := utils.WithCorrelationId(ctx.Request().Context(), runInput.CorrelationId.String())
		context = utils.WithAccount(context, runInput.Account)

		// take from the rate limit bucket
		this.rateLimiter.Wait(ctx.Request().Context())

		createError := sendToCloudConnector(
			false,
			runInput.Account,
			recipient,
			runInput,
			this.cloudConnectorClient,
			context,
		)

		if createError != nil {
			return createError
		}

		runId, recordCreateError := recordRunInformation(
			runInput,
			recipient,
			runInput.CorrelationId,
			false,
			false,
			requestType,
			this.database,
			this.config,
			context,
		)
		if recordCreateError != nil {
			return recordCreateError
		}

		return &RunCreated{
			Code: http.StatusCreated,
			Id:   runId,
		}
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}
