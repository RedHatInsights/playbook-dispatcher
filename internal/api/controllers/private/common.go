package private

import (
	"context"
	"playbook-dispatcher/internal/common/utils"
)

func (this *controllers) getEAN(ctx context.Context, orgID, recipient string) (ean *string, err error) {
	// TODO: temporary implementation until we have proper implementation
	if this.config.Get("tenant.translator.impl") == "impl" {
		var orgID string
		_, ean, err = this.translator.RHCIDToTenantIDs(ctx, recipient)
		utils.GetLogFromContext(ctx).Debugw("Received translated tenant info", "recipient", recipient, "org_id", orgID, "ean", ean)
	} else {
		ean, err = this.translator.OrgIDToEAN(ctx, orgID)
	}

	return
}
