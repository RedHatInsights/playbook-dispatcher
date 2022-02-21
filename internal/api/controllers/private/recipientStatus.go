package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/connectors/tenants"
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
			if _, ok := err.(*tenants.TenantNotFoundError); ok {
				return ctx.NoContent(http.StatusBadRequest)
			}

			utils.GetLogFromEcho(ctx).Error(err)
			return ctx.NoContent(http.StatusInternalServerError)
		}

		if ean == nil {
			utils.GetLogFromEcho(ctx).Warnw("Anemic tenant not supported", "org_id", string(recipient.OrgId))
			return ctx.NoContent(http.StatusBadRequest)
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

		results[i] = recipientStatusResponse(recipient, status == connectors.ConnectionStatus_connected)
	}

	return ctx.JSON(http.StatusOK, results)
}

func recipientStatusResponse(recipient RecipientWithOrg, connected bool) RecipientStatus {
	return RecipientStatus{
		RecipientWithOrg: recipient,
		Connected:        connected,
	}
}
