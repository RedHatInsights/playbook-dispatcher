package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
)

func (this *controllers) ApiInternalV2RunsCancel(ctx echo.Context) error {
	var input CancelationInputV2List

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// process individual requests concurrently
	result := input.PMapRunCanceled(func(cancelInputV2 CancelInputV2) *RunCanceled {
		context := utils.WithOrgId(ctx.Request().Context(), string(cancelInputV2.OrgId))
		context = utils.WithRequestType(context, instrumentation.LabelAnsibleRequest)

		parsedRunId := parseValidatedUUID(string(cancelInputV2.RunId))

		cancelInput := CancelInputV2GenericMap(cancelInputV2, parsedRunId)

		runID, _, err := this.dispatchManager.ProcessCancel(context, cancelInput.OrgId, cancelInput)
		if err != nil {
			return handleRunCancelError(err)
		}

		return runCanceled(runID)
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}
