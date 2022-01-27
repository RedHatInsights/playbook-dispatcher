package private

import (
	"fmt"
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

	// translate org_id to EAN for Cloud Connector
	// TODO: this will go away in the future
	orgToEAN := make(map[string]string)

	for _, recipient := range input {
		orgId := string(recipient.OrgId)
		if _, ok := orgToEAN[orgId]; !ok {
			ean, err := this.translator.OrgIDToEAN(ctx.Request().Context(), orgId)

			if err != nil {
				utils.GetLogFromEcho(ctx).Error(err)
				return ctx.NoContent(http.StatusInternalServerError)
			}

			if ean == nil {
				utils.GetLogFromEcho(ctx).Error(fmt.Errorf("Anemic tenant not supported: %s", orgId))
				return ctx.NoContent(http.StatusInternalServerError)
			}

			orgToEAN[orgId] = *ean
		}
	}

	// get connection status from Cloud Connector
	results := make([]RecipientStatus, len(input))
	for i, recipient := range input {
		account := orgToEAN[string(recipient.OrgId)]

		// take from the rate limit bucket
		// TODO: consider moving this to the httpClient level (e.g. as an HttpRequestDoer decorator)
		this.rateLimiter.Wait(ctx.Request().Context())

		// TODO: parallelize this
		status, err := this.cloudConnectorClient.GetConnectionStatus(ctx.Request().Context(), account, string(recipient.OrgId), string(recipient.Recipient))
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
