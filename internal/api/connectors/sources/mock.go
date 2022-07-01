package sources

import (
	"context"
	"fmt"
)

var (
	rhcId           string = "d415fc2d-9700-4e30-9621-6a410ccc92d8"
	name            string = "test"
	statusAvailable string = "available"
)

type mockImpl struct{}

func (*mockImpl) GetSourceConnectionDetails(ctx context.Context, sourceId string) (SourceConnectionStatus, error) {

	if sourceId == "07c9268f-6dc2-4e05-be57-d9d252a6bb47" {
		return SourceConnectionStatus{}, fmt.Errorf("RHCStatus Not Found")
	}

	if sourceId == "5d322fdb-1de4-4402-b383-30f0f66b0bc1" {
		return SourceConnectionStatus{}, fmt.Errorf("RHCStatus Bad Request")
	}

	if sourceId == "d3966054-5d45-45a5-a4b4-2f34ea4ae9e0" {
		return SourceConnectionStatus{}, fmt.Errorf("Source Bad Request")
	}

	response := SourceConnectionStatus{
		ID:                 sourceId,
		SourceName:         &name,
		RhcID:              &rhcId,
		AvailabilityStatus: &statusAvailable,
	}

	return response, nil
}

func NewMockSourcesClient() SourcesConnector {
	return &mockImpl{}
}
