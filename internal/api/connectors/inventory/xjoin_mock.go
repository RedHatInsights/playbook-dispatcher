package inventory

import (
	"context"
	"fmt"
)

type xjoinConnectorClientMock struct {
}

func NewXjoinClientMock() XjoinConnectorClient {
	return &xjoinConnectorClientMock{}
}

func (mock *xjoinConnectorClientMock) GetHostDetailsRequest(
	ctx context.Context,
	IDs []string,
	hostsFilter HostFilter,
	orderBy HOSTS_ORDER_BY,
	orderHow ORDER_DIR,
	limit int,
	offset int,
) (response []*HostDetailsResponse, err error) {
	if len(IDs) < 1 {
		return nil, fmt.Errorf("GetHostDetailsRequest Failed, No IDs provided for Request")
	}

	if IDs[0] == "6e7b6317-0a2d-4552-a2f2-b7da0aece49d" {
		return nil, fmt.Errorf("xjoin error")
	}

	resp := []*HostDetailsResponse{
		{
			ID:          "6e7b6317-0a2d-4552-a2f2-b7da0aece49d",
			DisplayName: "test1",
			Facts: map[string]interface{}{
				"satellite": map[string]interface{}{
					"satellite_instance_id": "33cd8e39-13bb-4d02-8316-84b850dc5163",
					"satellite_version":     "6.11",
				},
			},
			CanonicalFacts: map[string]interface{}{
				"fqdn": "fqdn.test2.rhel8.insights.local",
			},
		}, {
			ID:          "22cd8e39-13bb-4d02-8316-84b850dc5136",
			DisplayName: "test2",
			Facts: map[string]interface{}{
				"satellite": map[string]interface{}{
					"satellite_instance_id": "44cd8e39-13bb-4d02-8316-84b850dc5152",
					"satellite_version":     "6.11",
				},
			},
			CanonicalFacts: map[string]interface{}{
				"fqdn": "fqdn.test2.rhel8.insights.local",
			},
		},
	}

	return resp, nil
}

func (mock *xjoinConnectorClientMock) GetSystemProfileDetailsRequest(
	ctx context.Context,
	IDs []string,
	hostsFilter HostFilter,
	orderBy HOSTS_ORDER_BY,
	orderHow ORDER_DIR,
	limit int,
	offset int,
) (response []*SystemProfileDetailsResponse, err error) {
	if len(IDs) < 1 {
		return nil, fmt.Errorf("GetSystemProfileDetailsRequest Failed, No IDs provided for Request")
	}

	if IDs[0] == "6e7b6317-0a2d-4552-a2f2-b7da0aece49d" {
		return nil, fmt.Errorf("xjoin error")
	}

	resp := []*SystemProfileDetailsResponse{
		{
			ID: "6e7b6317-0a2d-4552-a2f2-b7da0aece49d",
			SystemProfileFacts: map[string]interface{}{
				"rhc_client_id": "33cd8e39-13bb-4d02-8316-84b850dc5136",
			},
		},
		{
			ID: "22cd8e39-13bb-4d02-8316-84b850dc5136",
			SystemProfileFacts: map[string]interface{}{
				"rhc_client_id": "22cd8e39-13bb-4d02-8316-84b850dc5136",
			},
		},
	}

	return resp, nil
}
