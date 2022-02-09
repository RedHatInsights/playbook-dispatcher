package tenants

import (
	"context"
	"fmt"
	"playbook-dispatcher/internal/common/utils"
)

type mockTenantIDTranslator struct {
	orgIDToEAN map[string]*string
	eanToOrgID map[string]string
}

func NewMockTenantIDTranslator() TenantIDTranslator {
	orgIDToEAN := map[string]*string{
		"5318290":  utils.StringRef("901578"),
		"12900172": utils.StringRef("6377882"),
		"14656001": utils.StringRef("7135271"),
		"11789772": utils.StringRef("6089719"),
		"3340851": utils.StringRef("0369233"),
	}

	return &mockTenantIDTranslator{
		orgIDToEAN: orgIDToEAN,
		eanToOrgID: inverseMap(orgIDToEAN),
	}
}

func (this *mockTenantIDTranslator) EANToOrgID(ctx context.Context, ean string) (orgId string, err error) {
	value, ok := this.eanToOrgID[ean]

	if !ok {
		return "", unsupportedError()
	}

	return value, nil
}

func (this *mockTenantIDTranslator) OrgIDToEAN(ctx context.Context, orgId string) (ean *string, err error) {
	value, ok := this.orgIDToEAN[orgId]

	if !ok {
		return nil, unsupportedError()
	}

	return value, nil
}

func (this *mockTenantIDTranslator) RHCIDToTenantIDs(ctx context.Context, rhcID string) (orgId string, ean *string, err error) {
	return "", nil, fmt.Errorf("not implemented")
}

func inverseMap(input map[string]*string) (result map[string]string) {
	result = make(map[string]string, len(input))
	for key, value := range input {
		if value != nil {
			result[*value] = key
		}
	}

	return
}

func unsupportedError() error {
	return fmt.Errorf("Unsupported operation (mock implementation)")
}
