package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//go:generate fungen -types RunInputV2,*RunCreated -methods PMap -package private -filename utils.v2.gen.go
func (this *controllers) ApiInternalV2RunsCreate(ctx echo.Context) error {
	var input RunInputV2List

	db.SetLog(this.database, utils.GetLogFromEcho(ctx))
	defer db.ClearLog(this.database)

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// send all requests to cloud connector concurrently
	result := input.PMapRunCreated(func(runInputV2 RunInputV2) *RunCreated {
		ansibleReq, satReq := CheckV2ReqFields(runInputV2)
		if !ansibleReq && !satReq {
			return runCreateError(http.StatusBadRequest)
		}
		if ansibleReq && satReq {
			ansibleReq, satReq = false, true
		}

		requestType := "ansible"
		if satReq {
			requestType = "satellite"
		}

		runInput := RunInputV2GenericMap(runInputV2, satReq, this.config)

		recipient, err := uuid.Parse(runInput.Recipient)
		if err != nil {
			instrumentation.InvalidRecipientId(ctx, runInput.Recipient, err, requestType)
			return runCreateError(http.StatusBadRequest)
		}

		context := utils.WithRequestType(ctx.Request().Context(), requestType)
		context = utils.WithCorrelationId(context, runInput.CorrelationId.String())
		context = utils.WithAccount(context, *runInput.OrgId)

		// take from the rate limit bucket
		this.rateLimiter.Wait(ctx.Request().Context())

		ean, err := this.translator.OrgIDToEAN(context, *runInput.OrgId)
		if err != nil {
			return runCreateError(http.StatusBadRequest)
		}
		runInput.Account = *ean

		createError := sendToCloudConnector(
			satReq,
			*ean,
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
			true,
			satReq,
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
