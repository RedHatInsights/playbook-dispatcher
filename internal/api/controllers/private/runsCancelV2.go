package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/common/utils"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"

	"github.com/labstack/echo/v4"
)

//go:generate fungen -types CancelInputV2,*RunCanceled -methods PMap -package private -filename utils.del.v2.gen.go
func (this *controllers) ApiInternalV2RunsCancel(ctx echo.Context) error {
	var input CancelInputV2List

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// process individual requests concurrently
	result := input.PMapRunCanceled(func(cancelInputV2 CancelInputV2) *RunCanceled {
		context := utils.WithOrgId(ctx.Request().Context(), string(cancelInputV2.OrgId))
		context = utils.WithRequestType(context, instrumentation.LabelAnsibleRequest)

		// translate org_id to EAN
		// TODO: this will go away in the future
		ean, err := this.translator.OrgIDToEAN(ctx.Request().Context(), string(cancelInputV2.OrgId))
		if err != nil {
			if _, ok := err.(*tenantid.TenantNotFoundError); ok {
				return runCancelError(http.StatusNotFound)
			}

			utils.GetLogFromEcho(ctx).Error(err)
			return runCancelError(http.StatusInternalServerError)
		}

		if ean == nil {
			utils.GetLogFromEcho(ctx).Warnw("Anemic tenant not supported", "org_id", string(cancelInputV2.OrgId))
			return runCancelError(http.StatusBadRequest)
		}

		parsedRunId := parseValidatedUUID(string(cancelInputV2.RunId))

		cancelInput := CancelInputV2GenericMap(cancelInputV2, *ean, parsedRunId)

		runID, _, err := this.dispatchManager.ProcessCancel(context, *ean, cancelInput)
		if err != nil {
			return handleRunCancelError(err)
		}

		return runCanceled(runID)
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}
