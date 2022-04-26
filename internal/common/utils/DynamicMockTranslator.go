package utils

import (
	"context"

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
	mockAccount := orgId[:5] + "-test"
	return &mockAccount, nil
}

func (this *dynamicMockTranslator) EANToOrgID(ctx context.Context, ean string) (orgId string, err error) {
	return ean + "-test", nil
}

func (this *dynamicMockTranslator) EANsToOrgIDs(ctx context.Context, eans []string) (results []tenantid.TranslationResult, err error) {
	results = make([]tenantid.TranslationResult, len(eans))

	for _, ean := range eans {
		orgId, err := this.EANToOrgID(ctx, ean)

		if err == nil {
			r := tenantid.TranslationResult{
				OrgID: orgId,
				EAN:   &ean,
				Err:   nil,
			}
			results = append(results, r)
		}
	}

	return results, nil
}

func (this *dynamicMockTranslator) OrgIDsToEANs(ctx context.Context, orgIDs []string) (results []tenantid.TranslationResult, err error) {
	results = make([]tenantid.TranslationResult, len(orgIDs))

	for _, orgID := range orgIDs {
		ean, err := this.OrgIDToEAN(ctx, orgID)

		if err == nil && ean != nil {
			r := tenantid.TranslationResult{
				OrgID: orgID,
				EAN:   ean,
				Err:   nil,
			}
			results = append(results, r)
		}
	}

	return results, nil
}
