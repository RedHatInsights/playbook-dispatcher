package inventory

import (
	"context"
	"fmt"
)

type inventoryConnectorMock struct {
}

func InventoryConnectorClientMock() InventoryConnector {
	return &inventoryConnectorMock{}
}

func (this *inventoryConnectorMock) GetHostDetails(
	ctx context.Context,
	IDs []string,
	orderBy string,
	orderHow string,
	limit int,
	offset int,
) (details []HostDetails, err error) {

	if IDs[0] == "0e97ad0d-8649-4ef1-a3aa-492024cc84bf" {
		return []HostDetails{}, fmt.Errorf("timeout")
	}

	hostDetails := HostDetails{
		ID:          "c484f980-ab8d-401b-90e7-aa1d4ccf8c0e",
		DisplayName: "test",
		Facts: map[string]interface{}{
			"satellite": "f3c7c15a-0092-4aa3-bf0b-41b6d3864d6d",
		},
		CanonicalFacts: map[string]interface{}{
			"fqdn": "test_name",
		},
	}

	hostDetailsList := []HostDetails{hostDetails}

	return hostDetailsList, nil
}

func (this *inventoryConnectorMock) GetSystemProfileDetails(
	ctx context.Context,
	IDs []string,
	orderBy string,
	orderHow string,
	limit int,
	offset int,
) (details []SystemProfileDetails, err error) {

	if IDs[0] == "0e97ad0d-8649-4ef1-a3aa-492024cc84bf" {
		return []SystemProfileDetails{}, fmt.Errorf("timeout")
	}

	profileDetails := SystemProfileDetails{
		ID: "b3286229-fea6-49e6-84ed-cb8b9fa2d94a",
		SystemProfileFacts: map[string]interface{}{
			"rhc_client_id": "db74645a-af59-4552-8146-5b846e0803c7",
			"owner_id":      "12345",
		},
	}

	profileDetailsList := []SystemProfileDetails{profileDetails}

	return profileDetailsList, nil
}
