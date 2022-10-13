package inventory

import (
	"context"
	"fmt"
)

type inventoryConnectorMock struct {
}

func NewInventoryClientMock() InventoryConnector {
	return &inventoryConnectorMock{}
}

func (this *inventoryConnectorMock) GetHostConnectionDetails(
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

	ownerID := "12345"
	satelliteInstanceID := "bd54e0e9-5310-45be-b107-fd7c96672ce5"
	satelliteOrgID := "5"
	satelliteVersion := "6.11"
	rhcClientID := "32af5948-301f-449a-a25b-ff34c83264a2"

	hostDetails := HostDetails{
		ID:                  "c484f980-ab8d-401b-90e7-aa1d4ccf8c0e",
		OwnerID:             &ownerID,
		SatelliteInstanceID: &satelliteInstanceID,
		SatelliteVersion:    &satelliteVersion,
		SatelliteOrgID:      &satelliteOrgID,
	}

	directConnectDetails := HostDetails{
		ID:          "fe30b997-c15a-44a9-89df-c236c3b5c540",
		OwnerID:     &ownerID,
		RHCClientID: &rhcClientID,
	}

	hostDetailsList := []HostDetails{hostDetails, directConnectDetails}

	return hostDetailsList, nil
}
