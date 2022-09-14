package inventory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/spf13/viper"
)

type XjoinConnectorClient interface {
	GetHostDetailsRequest(
		ctx context.Context,
		IDs []string,
		hostsFilter HostFilter,
		orderBy HOSTS_ORDER_BY,
		orderHow ORDER_DIR,
		limit int,
		offset int,
	) (response []*HostDetailsResponse, err error)

	GetSystemProfileDetailsRequest(
		ctx context.Context,
		IDs []string,
		hostsFilter HostFilter,
		orderBy HOSTS_ORDER_BY,
		orderHow ORDER_DIR,
		limit int,
		offset int,
	) (response []*SystemProfileDetailsResponse, err error)
}

type XjoinConnectorClientImpl struct {
	client XjoinConnectorClient //nolint unused
}

func createHostFilterList(IDs []string) []*HostFilter {
	hostFilterList := make([]*HostFilter, len(IDs))
	for i, ID := range IDs {
		hostFilterList[i] = &HostFilter{Id: FilterStringWithWildcard{Eq: ID}}
	}

	return hostFilterList
}

func createHostDetailResponseList(resp *getHostDetailsResponse) []*HostDetailsResponse {
	hostDetailResponseList := make([]*HostDetailsResponse, len(resp.Hosts.Data))
	for i, response := range resp.Hosts.Data {
		hostDetailResponseList[i] = &HostDetailsResponse{
			ID:             response.Id,
			Facts:          response.Facts,
			CanonicalFacts: response.Canonical_facts,
		}
	}

	return hostDetailResponseList
}

func createSystemProfileResponseList(resp *getSystemProfileFactsResponse) []*SystemProfileDetailsResponse {
	systemProfileResponseList := make([]*SystemProfileDetailsResponse, len(resp.Hosts.Data))
	for i, response := range resp.Hosts.Data {
		systemProfileResponseList[i] = &SystemProfileDetailsResponse{
			ID:                 response.Id,
			SystemProfileFacts: response.System_profile_facts,
		}
	}

	return systemProfileResponseList
}

func createHostsFilter(IDs []string) HostFilter {
	return HostFilter{
		OR: createHostFilterList(IDs),
	}
}

func NewXjoinClientWithRequestDoer(cfg *viper.Viper) graphql.Client {
	return graphql.NewClient(
		fmt.Sprintf("%s://%s:%s/graphql",
			cfg.GetString("xjoin.connector.scheme"),
			cfg.GetString("xjoin.connector.host"),
			cfg.GetString("xjoin.connector.port")),
		http.DefaultClient)
}

func (this *XjoinConnectorClientImpl) GetHostDetailsRequest(
	ctx context.Context,
	client graphql.Client,
	IDs []string,
	orderBy HOSTS_ORDER_BY,
	orderHow ORDER_DIR,
	pageSize int,
	limit int,
	offset int,
) (response []*HostDetailsResponse, err error) {
	if len(IDs) == 0 {
		return nil, fmt.Errorf("GetHostDetailsRequest Failed, No IDs provided for Request")
	}

	hostFilter := createHostsFilter(IDs)
	resp, err := getHostDetails(ctx, client, hostFilter, orderBy, orderHow, limit, offset)

	hostDetailsResponse := createHostDetailResponseList(resp)

	return hostDetailsResponse, err
}

func (this *XjoinConnectorClientImpl) GetSystemProfileDetailsRequest(
	ctx context.Context,
	client graphql.Client,
	IDs []string,
	orderBy HOSTS_ORDER_BY,
	orderHow ORDER_DIR,
	pageSize int,
	limit int,
	offset int,
) (response []*SystemProfileDetailsResponse, err error) {
	if len(IDs) == 0 {
		return nil, fmt.Errorf("GetSystemProfileDetailsRequest Failed, No IDs provided for Request")
	}

	hostFilter := createHostsFilter(IDs)
	resp, err := getSystemProfileFacts(ctx, client, hostFilter, orderBy, orderHow, limit, offset)

	systemProfileResponse := createSystemProfileResponseList(resp)

	return systemProfileResponse, err
}
