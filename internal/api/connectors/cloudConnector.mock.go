package connectors

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type cloudConnectorClientMock struct{}

func NewConnectorClientMock() CloudConnectorClient {
	return &cloudConnectorClientMock{}
}

func (this *cloudConnectorClientMock) SendCloudConnectorRequest(
	ctx context.Context,
	orgID string,
	recipient uuid.UUID,
	url *string,
	directive string,
	metadata map[string]string,
) (*uuid.UUID, bool, error) {
	if recipient.String() == "b5fbb740-5590-45a4-8240-89192dc49199" {
		return nil, true, nil
	}

	if recipient.String() == "b31955fb-3064-4f56-ae44-a1c488a28587" {
		return nil, false, fmt.Errorf("timeout")
	}

	if recipient.String() == "9200e4a3-c97c-4021-9856-82fa4673e8d2" && metadata["sat_id"] != "9274c274-a258-5d00-91fe-dbe0f7849cef" {
		return nil, false, fmt.Errorf("sat_id mismatch")
	}

	id := uuid.New()

	return &id, false, nil
}

func (this *cloudConnectorClientMock) GetConnectionStatus(
	ctx context.Context,
	orgID string,
	recipient string,
) (ConnectionStatus, error) {
	if orgID == "5318290" && recipient == "411cb203-f8c9-480e-ba20-1efbc74e3a33" {
		return Disconnected, nil
	}

	return Connected, nil
}
