package sources

import "context"

type SourceConnectionStatus struct {
	ID                 string  `json:"id"`
	SourceName         *string `json:"name,omitempty"`
	RhcID              *string `json:"rhc_id,omitempty"`
	AvailabilityStatus *string `json:"availability_status,omitempty"`
}

type SourcesConnector interface {
	GetSourceConnectionDetails(ctx context.Context, ID string) (SourceConnectionStatus, error)
}
