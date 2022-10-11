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

	satelliteInstanceID := "bd54e0e9-5310-45be-b107-fd7c96672ce5"
	satelliteVersion := "6.11"
	rhcClientID := "32af5948-301f-449a-a25b-ff34c83264a2"

	hostDetails := HostDetails{
		ID:                  "c484f980-ab8d-401b-90e7-aa1d4ccf8c0e",
		OwnerID:             "12345",
		SatelliteInstanceID: &satelliteInstanceID,
		SatelliteVersion:    &satelliteVersion,
		RHCConnectionID:     &rhcClientID,
	}

	hostDetailsList := []HostDetails{hostDetails}

	return hostDetailsList, nil
}
