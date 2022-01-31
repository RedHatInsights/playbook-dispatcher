package tenants

import (
	"context"
)

// This service handles translation between tenant identifiers.
// Namely, it converts an org_id to EAN (EBS account number) and vice versa.
type TenantIDTranslator interface {

	// Converts an EAN (EBS account number) to org_id
	EANToOrgID(ctx context.Context, ean string) (orgId string, err error)

	// Converts an org_id to EAN (EBS account number). May return nil if the org_id belongs to an anemic tenant
	OrgIDToEAN(ctx context.Context, orgId string) (ean *string, err error)
}
