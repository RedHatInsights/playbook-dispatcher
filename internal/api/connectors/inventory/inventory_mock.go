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

	// if IDs[0] == "b5fbb740-5590-45a4-8240-89192dc49199" {
	// 	return nil, true, nil
	// }

	// if recipient.String() == "b31955fb-3064-4f56-ae44-a1c488a28587" {
	// 	return nil, false, fmt.Errorf("timeout")
	// }

	// if recipient.String() == "9200e4a3-c97c-4021-9856-82fa4673e8d2" && metadata["sat_id"] != "9274c274-a258-5d00-91fe-dbe0f7849cef" {
	// 	return nil, false, fmt.Errorf("sat_id mismatch")
	// }

	// id := uuid.New().String()

	// return &id, false, nil

	// return hostDetails, err
	fmt.Print("MOCK")
	hostDetails := HostDetails{
		ID:          "1234",
		DisplayName: "test",
		Facts: map[string]interface{}{
			"satellite": "blah",
		},
		CanonicalFacts: map[string]interface{}{
			"fqdn": "test",
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

	// ADD TESTING SHIT HERE
	profileDetails := SystemProfileDetails{
		ID: "5678",
		SystemProfileFacts: map[string]interface{}{
			"rhc_client_id": "1234",
		},
	}

	profileDetailsList := []SystemProfileDetails{profileDetails}

	return profileDetailsList, nil
}
