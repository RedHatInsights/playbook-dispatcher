package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
)

//go:generate fungen -types RunInput,*RunCreated -methods PMap -package private -filename utils.gen.go
func (this *controllers) ApiInternalRunsCreate(ctx echo.Context) error {
	var input RunInputList

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// process individual requests concurrently
	result := input.PMapRunCreated(func(runInputV1 RunInput) *RunCreated {
		context := utils.WithAccount(ctx.Request().Context(), string(runInputV1.Account))
		context = utils.WithRequestType(context, instrumentation.LabelAnsibleRequest)

		recipient := parseValidatedUUID(string(runInputV1.Recipient))
		hosts := parseRunHosts(runInputV1.Hosts)

		runInput := RunInputV1GenericMap(runInputV1, recipient, hosts, this.config)

		runID, _, err := this.dispatchManager.ProcessRun(context, string(runInputV1.Account), middleware.GetPSKPrincipal(context), runInput)

		if err != nil {
			return handleRunCreateError(err)
		}

		return runCreated(runID)
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}
