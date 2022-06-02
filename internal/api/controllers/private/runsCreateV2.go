package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/utils"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//go:generate fungen -types RunInputV2,*RunCreated -methods PMap -package private -filename utils.v2.gen.go
func (this *controllers) ApiInternalV2RunsCreate(ctx echo.Context) error {
	var input RunInputV2List

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	for _, run := range input {
		err = validateSatelliteFields(run)
		if err != nil {
			instrumentation.InvalidSatelliteRequest(ctx, err)
			return invalidRequest(ctx, err)
		}
	}

	// process individual requests concurrently
	result := input.PMapRunCreated(func(runInputV2 RunInputV2) *RunCreated {
		context := utils.WithOrgId(ctx.Request().Context(), string(runInputV2.OrgId))
		context = utils.WithRequestType(context, getRequestTypeLabel(runInputV2))

		recipient := parseValidatedUUID(string(runInputV2.Recipient))

		// translate org_id to EAN
		ean, err := this.translator.OrgIDToEAN(ctx.Request().Context(), string(runInputV2.OrgId))
		if err != nil {
			if _, ok := err.(*tenantid.TenantNotFoundError); ok {
				return runCreateError(http.StatusNotFound)
			}

			utils.GetLogFromEcho(ctx).Error(err)
			return runCreateError(http.StatusInternalServerError)
		}

		if ean == nil {
			utils.GetLogFromEcho(ctx).Warnw("Anemic tenant not supported", "org_id", string(runInputV2.OrgId))
			return runCreateError(http.StatusBadRequest)
		}

		hosts := parseRunHosts(runInputV2.Hosts)

		var parsedSatID *uuid.UUID
		if runInputV2.RecipientConfig != nil && runInputV2.RecipientConfig.SatId != nil {
			parsedSatID = utils.UUIDRef(parseValidatedUUID(string(*runInputV2.RecipientConfig.SatId)))
		}

		runInput := RunInputV2GenericMap(runInputV2, *ean, recipient, hosts, parsedSatID, this.config)

		runID, _, err := this.dispatchManager.ProcessRun(context, *ean, middleware.GetPSKPrincipal(context), runInput)

		if err != nil {
			return handleRunCreateError(err)
		}

		return runCreated(runID)
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}

func getRequestTypeLabel(run RunInputV2) string {
	result := instrumentation.LabelAnsibleRequest

	if run.RecipientConfig != nil && run.RecipientConfig.SatId != nil {
		result = instrumentation.LabelSatRequest
	}

	return result
}
