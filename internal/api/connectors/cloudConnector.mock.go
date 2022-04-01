package connectors

import (
	"context"
	"fmt"

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
	url *string,
	directive string,
	metadata map[string]string,

) (*string, bool, error) {
	if recipient.String() == "b5fbb740-5590-45a4-8240-89192dc49199" {
		return nil, true, nil
	}

	if recipient.String() == "b31955fb-3064-4f56-ae44-a1c488a28587" {
		return nil, false, fmt.Errorf("timeout")
	}

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
