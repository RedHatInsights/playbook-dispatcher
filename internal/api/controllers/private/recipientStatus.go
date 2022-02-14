package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
)

func (this *controllers) ApiInternalV2RecipientsStatus(ctx echo.Context) error {
	var input []RecipientWithOrg

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// get connection status from Cloud Connector
	results := make([]RecipientStatus, len(input))
	for i, recipient := range input {

		// translate org_id to EAN for Cloud Connector
		// TODO: this will go away in the future
		ean, err := this.getEAN(ctx.Request().Context(), string(recipient.OrgId), string(recipient.Recipient))
		if err != nil {
			utils.GetLogFromEcho(ctx).Error(err)
			return ctx.NoContent(http.StatusInternalServerError)
		}

		if ean == nil {
			utils.GetLogFromEcho(ctx).Infow("No EAN for org_id/recipient combination", "recipient", string(recipient.Recipient), "org_id", string(recipient.OrgId))
			results[i] = RecipientStatus{
				RecipientWithOrg: recipient,
				Connected:        false,
			}

			continue
		}

		// take from the rate limit bucket
		// TODO: consider moving this to the httpClient level (e.g. as an HttpRequestDoer decorator)
		this.rateLimiter.Wait(ctx.Request().Context())

		// TODO: parallelize this
		status, err := this.cloudConnectorClient.GetConnectionStatus(ctx.Request().Context(), *ean, string(recipient.OrgId), string(recipient.Recipient))
		if err != nil {
			utils.GetLogFromEcho(ctx).Error(err)
			return ctx.NoContent(http.StatusInternalServerError)
		}

		results[i] = RecipientStatus{
			RecipientWithOrg: recipient,
			Connected:        status == connectors.ConnectionStatus_connected,
		}
	}

	return ctx.JSON(http.StatusOK, results)
}
