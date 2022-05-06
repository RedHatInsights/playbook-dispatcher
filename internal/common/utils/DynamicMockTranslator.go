package utils

import (
	"context"
	"fmt"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"
)

type dynamicMockTranslator struct {
	tenantid.BatchTranslator
}

func NewDynamicMockTranslator() tenantid.Translator {
	return &dynamicMockTranslator{
		BatchTranslator: &dynamicMockTranslator{},
	}
}

func (this *dynamicMockTranslator) OrgIDToEAN(ctx context.Context, orgId string) (ean *string, err error) {
	mockAccount := fmt.Sprintf("%05s", orgId)[:5] + "-test"

	orgIdsWithNoAccount := []string{"1234", "abcd", "654321", "654322"}
	for _, orgIdWithNoAccount := range orgIdsWithNoAccount {
		if orgId == orgIdWithNoAccount {
			return nil, nil
		}
	}

	return &mockAccount, nil
}

func (this *dynamicMockTranslator) EANToOrgID(ctx context.Context, ean string) (orgId string, err error) {

	eanWithNoOrgId := []string{"1234", "abcd"}
	for _, eanWithNoOrgId := range eanWithNoOrgId {
		if ean == eanWithNoOrgId {
			return "", &tenantid.TenantNotFoundError{}
		}
	}

	return ean + "-test", nil
}

func (this *dynamicMockTranslator) EANsToOrgIDs(ctx context.Context, eans []string) (results []tenantid.TranslationResult, err error) {
	results = make([]tenantid.TranslationResult, len(eans))

	for _, ean := range eans {
		orgId, err := this.EANToOrgID(ctx, ean)

		r := tenantid.TranslationResult{
			OrgID: orgId,
			EAN:   &ean,
			Err:   err,
		}

		results = append(results, r)
	}

	return results, nil
}

func (this *dynamicMockTranslator) OrgIDsToEANs(ctx context.Context, orgIDs []string) (results []tenantid.TranslationResult, err error) {
	results = make([]tenantid.TranslationResult, len(orgIDs))

	for _, orgID := range orgIDs {
		ean, err := this.OrgIDToEAN(ctx, orgID)

		r := tenantid.TranslationResult{
			OrgID: orgID,
			EAN:   ean,
			Err:   err,
		}

		results = append(results, r)
	}

	return results, nil
}
