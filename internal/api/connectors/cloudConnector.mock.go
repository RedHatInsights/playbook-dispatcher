package connectors

import (
	"context"

	"github.com/google/uuid"
)

type cloudConnectorClientMock struct {
}

func NewConnectorClientMock() CloudConnectorClient {
	return &cloudConnectorClientMock{}
}

func (this *cloudConnectorClientMock) SendCloudConnectorRequest(
	ctx context.Context,
	account string,
	recipient uuid.UUID,
	correlationId uuid.UUID,
	url string,
) (*string, bool, error) {
	id := uuid.New().String()
	return &id, false, nil
}

func (this *cloudConnectorClientMock) GetConnectionStatus(
	ctx context.Context,
	account string,
	orgID string,
	recipient string,
) (ConnectionStatus, error) {
	if orgID == "5318290" && recipient == "411cb203-f8c9-480e-ba20-1efbc74e3a33" {
		return ConnectionStatus_disconnected, nil
	}

	return ConnectionStatus_connected, nil
}
