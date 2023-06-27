package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/model/generic"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (this *controllers) ApiInternalHighlevelJobRequest(ctx echo.Context) error {
	var input JobRequestBody
	var response []JobRequestInfo

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	hostConnectorDetails, err := this.inventoryConnectorClient.GetHostConnectionDetails(
		ctx.Request().Context(),
		input.Hosts,
		this.config.GetString("inventory.connector.ordered.how"),
		this.config.GetString("inventory.connector.ordered.by"),
		this.config.GetInt("inventory.connector.limit"),
		this.config.GetInt("inventory.connector.offset"),
	)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	satellite, directConnected, noRhc := sortHostsByRecipient(hostConnectorDetails)

	runInputs := []generic.RunInput{}
	runInputConnInfo := []RecipientWithConnectionInfo{}

	// when RHC is not setup, do not dispatch playbook run request
	if noRhc != nil {
		noRHCRecipientConnInfo := []RecipientWithConnectionInfo{getRHCStatus(noRhc, input.OrgId)}
		for _, noRHCRecipientConn := range noRHCRecipientConnInfo {
			response = append(response, JobRequestInfo{
				Recipient:       noRHCRecipientConn.Recipient,
				Status:          noRHCRecipientConn.Status,
				RequestDispatch: "failure",
				Systems:         noRHCRecipientConn.Systems,
			})
		}
	}

	if len(satellite) > 0 {
		satelliteStatuses, err := getSatelliteStatus(ctx, this.cloudConnectorClient, this.sourcesConnectorClient, input.OrgId, satellite)

		if err != nil {
			utils.GetLogFromEcho(ctx).Errorf("Error retrieving Satellite status: %s", err)
		}
		for _, satelliteStatus := range satelliteStatuses {
			if satelliteStatus.Status != "connected" {
				response = append(response, JobRequestInfo{
					Recipient:       satelliteStatus.Recipient,
					Status:          satelliteStatus.Status,
					RequestDispatch: "failure",
					Systems:         satelliteStatus.Systems,
				})
			} else {
				runInputs = append(runInputs, runInputFormatter(input, satelliteStatus))
				runInputConnInfo = append(runInputConnInfo, satelliteStatus)
			}
		}
	}

	if len(directConnected) > 0 {
		directConnectedStatuses, err := getDirectConnectStatus(ctx, this.cloudConnectorClient, input.OrgId, directConnected)

		if err != nil {
			utils.GetLogFromEcho(ctx).Errorf("Error retrieving Direct Connect status: %s", err)
		}
		for _, directConnectedStatus := range directConnectedStatuses {
			if directConnectedStatus.Status != "connected" {
				response = append(response, JobRequestInfo{
					Recipient:       directConnectedStatus.Recipient,
					Status:          directConnectedStatus.Status,
					RequestDispatch: "failure",
					Systems:         directConnectedStatus.Systems,
				})
			} else {
				runInputs = append(runInputs, runInputFormatter(input, directConnectedStatus))
				runInputConnInfo = append(runInputConnInfo, directConnectedStatus)
			}
		}
	}

	for i, runInput := range runInputs {
		context := ctx.Request().Context()
		runID, _, err := this.dispatchManager.ProcessRun(context, runInput.OrgId, middleware.GetPSKPrincipal(context), runInput)

		if err != nil {
			utils.GetLogFromEcho(ctx).Warnf("Error processing playbook run: %s", err)
			response = append(response, JobRequestInfo{
				Recipient:       runInputConnInfo[i].Recipient,
				Status:          runInputConnInfo[i].Status,
				RequestDispatch: "failure",
				Systems:         runInputConnInfo[i].Systems,
			})
			continue
		}

		convertedRunId := public.RunId(runID.String())
		response = append(response, JobRequestInfo{
			Recipient:       runInputConnInfo[i].Recipient,
			Status:          runInputConnInfo[i].Status,
			RequestDispatch: "success",
			Systems:         runInputConnInfo[i].Systems,
			RunId:           &convertedRunId,
		})
	}

	return ctx.JSON(http.StatusOK, HighLevelJobRequestResponse(response))
}

func runInputFormatter(input JobRequestBody, hostStatus RecipientWithConnectionInfo) generic.RunInput {
	var runInput generic.RunInput
	runInput.OrgId = string(hostStatus.OrgId)
	runInput.Principal = (*string)(&input.Principal)
	runInput.Recipient, _ = uuid.Parse(string(hostStatus.Recipient))
	runInput.Url = string(input.Url)
	runInput.Name = (*string)(&input.PlaybookName)

	var runInputHosts []generic.RunHostsInput
	for _, hostId := range hostStatus.Systems {
		parsedHostId, _ := uuid.Parse(string(hostId))
		runInputHosts = append(runInputHosts, generic.RunHostsInput{InventoryId: &parsedHostId})
	}
	runInput.Hosts = runInputHosts

	if hostStatus.RecipientType == "satellite" {
		hostSatId, _ := uuid.Parse(string(hostStatus.SatId))
		runInput.SatId = &hostSatId

		satOrgId := string(hostStatus.SatOrgId)
		runInput.SatOrgId = &satOrgId
	}

	if input.WebConsoleUrl != nil {
		runInput.WebConsoleUrl = (*string)(input.WebConsoleUrl)
	}
	if input.Timeout != nil {
		runInput.Timeout = (*int)(input.Timeout)
	}
	if input.Labels != nil {
		runInput.Labels = input.Labels.AdditionalProperties
	}

	return runInput
}
