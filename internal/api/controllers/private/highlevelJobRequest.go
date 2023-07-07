package private

import (
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (this *controllers) ApiInternalHighlevelJobRequest(ctx echo.Context) error {
	var input RunInputV2List

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	for _, run := range input {
		requestType, err := validateJobRequestFields(run)
		if err != nil {
			instrumentation.InvalidHighLevelJobRequest(ctx, requestType, err)
			return ctx.NoContent(http.StatusBadRequest)
		}
	}

	// processing individual requests concurrently
	result := input.PMapRunCreated(func(runInputV2 RunInputV2) *RunCreated {
		requestTypeLabel := getRequestTypeLabel(runInputV2)
		context := utils.WithOrgId(ctx.Request().Context(), string(runInputV2.OrgId))
		context = utils.WithRequestType(context, requestTypeLabel)

		recipient := parseValidatedUUID(string(runInputV2.Recipient))

		hosts := parseRunHosts(runInputV2.Hosts)

		var parsedSatID *uuid.UUID
		if runInputV2.RecipientConfig != nil && runInputV2.RecipientConfig.SatId != nil {
			parsedSatID = utils.UUIDRef(parseValidatedUUID(string(*runInputV2.RecipientConfig.SatId)))
		}

		runInput := RunInputV2GenericMap(runInputV2, recipient, hosts, parsedSatID, this.config)

		hostConnectorDetails, err := this.inventoryConnectorClient.GetHostConnectionDetails(
			context,
			extractInventoryIds(runInput.Hosts),
			this.config.GetString("inventory.connector.ordered.how"),
			this.config.GetString("inventory.connector.ordered.by"),
			this.config.GetInt("inventory.connector.limit"),
			this.config.GetInt("inventory.connector.offset"),
		)
		if err != nil {
			utils.GetLogFromContext(context).Error(err)
			return handleRunCreateError(err)
		}

		satellite, directConnected, noRhc := sortHostsByRecipient(hostConnectorDetails)

		if len(satellite) > 1 || len(directConnected) > 1 || len(noRhc) > 1 {
			instrumentation.InvalidHighLevelJobRequest(ctx, requestTypeLabel, fmt.Errorf("Multiple recipients for host list found"))
			return &RunCreated{Code: http.StatusBadRequest}
		}

		// when RHC is not setup, do not dispatch playbook run request
		if noRhc != nil {
			noRHCRecipientConnInfo := []RecipientWithConnectionInfo{getRHCStatus(noRhc, OrgId(runInput.OrgId))}
			instrumentation.InvalidHighLevelJobRequest(ctx, requestTypeLabel, fmt.Errorf("RHC not installed. Recipient: %s", noRHCRecipientConnInfo[0].Recipient))
			return runCreateError(http.StatusNotFound)
		}

		if satellite != nil {
			satelliteStatus, err := getSatelliteStatus(ctx, this.cloudConnectorClient, this.sourcesConnectorClient, OrgId(runInput.OrgId), satellite)
			if err != nil {
				utils.GetLogFromContext(context).Errorf("Error retrieving Satellite status: %s", err)
				return runCreateError(http.StatusNotFound)
			}
			if satelliteStatus[0].Status != "connected" {
				instrumentation.InvalidHighLevelJobRequest(ctx, requestTypeLabel, fmt.Errorf("Satellite not connected"))
				return runCreateError(http.StatusNotFound)
			}
		}

		if directConnected != nil {
			directConnectedStatus, err := getDirectConnectStatus(ctx, this.cloudConnectorClient, OrgId(runInput.OrgId), directConnected)
			if err != nil {
				utils.GetLogFromContext(context).Errorf("Error Retrieving Direct Connect status: %s", err)
				return runCreateError(http.StatusNotFound)
			}
			if directConnectedStatus[0].Status != "connected" {
				instrumentation.InvalidHighLevelJobRequest(ctx, requestTypeLabel, fmt.Errorf("Host not connected"))
				return runCreateError(http.StatusNotFound)
			}
		}

		runID, _, err := this.dispatchManager.ProcessRun(context, runInput.OrgId, middleware.GetPSKPrincipal(context), runInput)

		if err != nil {
			return handleRunCreateError(err)
		}

		return runCreated(runID)
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}
