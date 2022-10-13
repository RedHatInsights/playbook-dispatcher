package private

import (
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/connectors/inventory"
	"playbook-dispatcher/internal/api/connectors/sources"
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
)

type rhcSatellite struct {
	SatelliteInstanceID      string
	SatelliteOrgID           string
	SatelliteVersion         string
	Hosts                    []string
	SourceID                 string
	RhcClientID              *string
	SourceAvailabilityStatus *string
}

func (this *controllers) ApiInternalHighlevelConnectionStatus(ctx echo.Context) error {
	var input HostsWithOrgId
	satelliteResponses := []RecipientWithConnectionInfo{}
	directConnectedResponses := []RecipientWithConnectionInfo{}
	noRHCResponses := []RecipientWithConnectionInfo{}

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
		return ctx.NoContent(http.StatusBadRequest)
	}

	satellite, directConnected, noRhc := sortHostsByRecipient(hostConnectorDetails)

	// Return noRHC If no Satellite or Direct Connected hosts exist
	if noRhc != nil {
		noRHCResponses = []RecipientWithConnectionInfo{getRHCStatus(noRhc, input.OrgId)}
	}

	if satellite == nil && directConnected == nil {
		return ctx.JSON(http.StatusAccepted, noRHCResponses)
	}

	if len(satellite) > 0 {
		satelliteResponses, err = getSatelliteStatus(ctx, this.cloudConnectorClient, this.sourcesConnectorClient, input.OrgId, satellite)

		if err != nil {
			return fmt.Errorf("Error retrieving Satellite status: %s", err)
		}
	}

	if len(directConnected) > 0 {
		directConnectedResponses, err = getDirectConnectStatus(ctx, this.cloudConnectorClient, input.OrgId, directConnected)

		if err != nil {
			return fmt.Errorf("Error retrieving DirectConnect status: %s", err)
		}
	}

	return ctx.JSON(http.StatusOK, HighLevelRecipientStatus(concatResponses(satelliteResponses, directConnectedResponses, noRHCResponses)))
}

func sortHostsByRecipient(details []inventory.HostDetails) (satelliteDetails []inventory.HostDetails, directConnectedDetails []inventory.HostDetails, noRhc []inventory.HostDetails) {
	recipients := make(map[int][]inventory.HostDetails, 3)

	for _, host := range details {
		switch {
		case host.SatelliteInstanceID != nil:
			recipients[0] = append(recipients[0], host) // If satellite_instance_id exitsts Satellite host
		case host.RHCClientID != nil:
			recipients[1] = append(recipients[1], host) // if rhc_client_id exists in inventory facts host is direct connect
		default:
			recipients[2] = append(recipients[2], host)
		}
	}

	return recipients[0], recipients[1], recipients[2]
}

func formatConnectionResponse(satID *string, satOrgID *string, rhcClientID *string, orgID OrgId, hosts []string, recipientType string, status string) RecipientWithConnectionInfo {
	formatedHosts := make([]HostId, len(hosts))
	var formatedSatID SatelliteId
	var formatedSatOrgID SatelliteOrgId
	var formatedRHCClientID public.RunRecipient

	if satID != nil {
		formatedSatID = SatelliteId(*satID)
	}

	if satOrgID != nil {
		formatedSatOrgID = SatelliteOrgId(*satOrgID)
	}

	if rhcClientID != nil {
		formatedRHCClientID = public.RunRecipient(*rhcClientID)
	}

	for i, host := range hosts {
		formatedHosts[i] = HostId(host)
	}

	connectionInfo := RecipientWithConnectionInfo{
		OrgId:         orgID,
		Recipient:     formatedRHCClientID,
		RecipientType: RecipientType(recipientType),
		SatId:         formatedSatID,
		SatOrgId:      formatedSatOrgID,
		Status:        status,
		Systems:       formatedHosts,
	}

	return connectionInfo
}

func getDirectConnectStatus(ctx echo.Context, client connectors.CloudConnectorClient, orgId OrgId, hostDetails []inventory.HostDetails) ([]RecipientWithConnectionInfo, error) {
	responses := []RecipientWithConnectionInfo{}
	for _, host := range hostDetails {
		status, err := client.GetConnectionStatus(ctx.Request().Context(), string(orgId), *host.RHCClientID)

		if err != nil {
			utils.GetLogFromEcho(ctx).Error(err)
			return nil, ctx.NoContent(http.StatusInternalServerError)
		}

		recipient := RecipientWithOrg{
			OrgId:     orgId,
			Recipient: public.RunRecipient(*host.RHCClientID),
		}

		results := recipientStatusResponse(recipient, status == connectors.ConnectionStatus_connected)

		var connectionStatus string
		if results.Connected {
			connectionStatus = "connected"
		} else {
			connectionStatus = "disconnected"
		}

		responses = append(responses, formatConnectionResponse(nil, nil, host.RHCClientID, orgId, []string{host.ID}, string(RecipientType_directConnect), connectionStatus))
	}

	return responses, nil
}

func getSatelliteStatus(ctx echo.Context, client connectors.CloudConnectorClient, sourceClient sources.SourcesConnector, orgId OrgId, hostDetails []inventory.HostDetails) ([]RecipientWithConnectionInfo, error) {
	hostsGroupedBySatellite := make(map[string]*rhcSatellite)

	// Group hosts by Satellite (instance_id + org_id)
	for _, host := range hostDetails {
		satInstanceAndOrg := *host.SatelliteInstanceID + *host.SatelliteOrgID
		_, exists := hostsGroupedBySatellite[satInstanceAndOrg]

		if exists {
			hostsGroupedBySatellite[satInstanceAndOrg].Hosts = append(hostsGroupedBySatellite[satInstanceAndOrg].Hosts, host.ID)
		} else {
			hostsGroupedBySatellite[satInstanceAndOrg] = &rhcSatellite{
				SatelliteInstanceID: *host.SatelliteInstanceID,
				SatelliteOrgID:      *host.SatelliteOrgID,
				SatelliteVersion:    *host.SatelliteVersion,
				Hosts:               []string{host.ID},
			}
		}
	}

	// Get source info
	for i, satellite := range hostsGroupedBySatellite {
		result, err := sourceClient.GetSourceConnectionDetails(ctx.Request().Context(), satellite.SatelliteInstanceID)

		if err != nil {
			return nil, ctx.NoContent(http.StatusInternalServerError)
		}

		hostsGroupedBySatellite[i].SourceID = result.ID
		hostsGroupedBySatellite[i].RhcClientID = result.RhcID
		hostsGroupedBySatellite[i].SourceAvailabilityStatus = result.AvailabilityStatus
	}

	// Get Connection Status
	responses := []RecipientWithConnectionInfo{}
	for _, satellite := range hostsGroupedBySatellite {
		if satellite.RhcClientID != nil {
			status, err := client.GetConnectionStatus(ctx.Request().Context(), satellite.SatelliteOrgID, satellite.SatelliteInstanceID)
			if err != nil {
				utils.GetLogFromEcho(ctx).Error(err)
				return nil, ctx.NoContent(http.StatusInternalServerError)
			}

			recipient := RecipientWithOrg{
				OrgId:     orgId,
				Recipient: public.RunRecipient(*satellite.RhcClientID),
			}

			results := recipientStatusResponse(recipient, status == connectors.ConnectionStatus_connected)

			var connectionStatus string
			if results.Connected {
				connectionStatus = "connected"
			} else {
				connectionStatus = "disconnected"
			}

			responses = append(responses, formatConnectionResponse(&satellite.SourceID, &satellite.SatelliteOrgID, satellite.RhcClientID, orgId, satellite.Hosts, string(RecipientType_satellite), connectionStatus))
		}
	}

	return responses, nil
}

func getRHCStatus(hostDetails []inventory.HostDetails, orgID OrgId) RecipientWithConnectionInfo {
	hostIDs := make([]string, len(hostDetails))

	for i, host := range hostDetails {
		hostIDs[i] = host.ID
	}

	return formatConnectionResponse(nil, nil, nil, orgID, hostIDs, "none", "no_rhc")
}

func concatResponses(satellite []RecipientWithConnectionInfo, directConnect []RecipientWithConnectionInfo, noRHC []RecipientWithConnectionInfo) []RecipientWithConnectionInfo {
	responses := append(satellite, directConnect...)

	return append(responses, noRHC...)
}
