package tenants

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type translatorClientImpl struct {
	client      operationAwareDoer
	serviceHost string
}

func (this *translatorClientImpl) EANToOrgID(ctx context.Context, ean string) (orgId string, err error) {
	return "", fmt.Errorf("Not implemented")
}

func (this *translatorClientImpl) OrgIDToEAN(ctx context.Context, orgId string) (ean *string, err error) {
	return nil, fmt.Errorf("Not implemented")
}

func (this *translatorClientImpl) RHCIDToTenantIDs(ctx context.Context, rhcID string) (orgId string, ean *string, err error) {
	url := fmt.Sprintf("%s/internal/certauth", this.serviceHost)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-rh-certauth-cn", fmt.Sprintf("/CN=%s", rhcID))

	r, err := this.client.Do(req, "RHCIDToTenantIDs")
	if err != nil {
		return "", nil, err
	}

	defer r.Body.Close()

	if r.StatusCode == 404 {
		return "", nil, &TenantNotFoundError{
			msg: fmt.Sprintf("Tenant not found. RHCID: %s", rhcID),
		}
	}
	if r.StatusCode != 200 {
		return "", nil, fmt.Errorf("Unexpected status code %d for id %s", r.StatusCode, rhcID)
	}

	var resp authGwResp
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return "", nil, err
	}

	idRaw, err := base64.StdEncoding.DecodeString(resp.Identity)
	if err != nil {
		return "", nil, err
	}

	var jsonData xrhid
	err = json.Unmarshal(idRaw, &jsonData)
	if err != nil {
		return "", nil, err
	}

	return jsonData.Identity.OrgID, &jsonData.Identity.AccountNumber, nil
}

type authGwResp struct {
	Identity string `json:"x-rh-identity"`
}

type xrhid struct {
	Identity identity `json:"identity"`
}

type identity struct {
	AccountNumber string `json:"account_number"`
	OrgID         string `json:"org_id"`
}
