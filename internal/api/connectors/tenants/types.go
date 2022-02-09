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

	// Converts an RHC ID to org_id and EAN (EBS account number). EAN may be nil if the host belongs to an anemic tenant
	RHCIDToTenantIDs(ctx context.Context, rhcID string) (orgId string, ean *string, err error)
}

// Indicates that no tenant matches the provided identifier
type TenantNotFoundError struct {
	msg string
}

func (e *TenantNotFoundError) Error() string { return e.msg }

type TenantIDTranslatorOption interface {
	apply(*translatorClientImpl)
	// Options may need to be applied in certain order (e.g. timeout-setting should be applied before an option that wraps the doer)
	// Priority lets us have complete ordering of all options
	// Options are applied in ascending priority (highest priority options go last)
	priority() int
}
