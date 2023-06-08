package inventory

import (
	"context"
)

type satelliteFacts struct {
	SatelliteInstanceID *string `json:"satellite_instance_id,omitempty"`
	SatelliteVersion    *string `json:"satellite_version,omitempty"`
	SatelliteOrgID      *string `json:"satellite_org_id,omitempty"`
}

type HostDetails struct {
	ID                  string  `json:"id"`
	OwnerID             *string `json:"owner_id,omitempty"`
	SatelliteInstanceID *string `json:"satellite_instance_id,omitempty"`
	SatelliteVersion    *string `json:"satellite_version,omitempty"`
	SatelliteOrgID      *string `json:"satellite_org_id,omitempty"`
	RHCClientID         *string `json:"rhc_client_id,omitempty"`
}

type InventoryConnector interface {
	GetHostConnectionDetails(ctx context.Context, IDs []string, order_how string, order_by string, limit int, offset int) ([]HostDetails, error)
}
