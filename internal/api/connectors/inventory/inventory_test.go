package inventory

import (
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var (
	orderBy  = HOSTS_ORDER_BY("display_name")
	orderHow = ORDER_DIR("ASC")
	limit    = 10
	offset   = 0
)

var _ = Describe("Xjoin", func() {
	Describe("GetHostDetailsRequest", func() {
		It("Interperates response correctly", func() {
			client := NewXjoinClientMock()

			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"1234", "5678"}
			hostFilter := createHostsFilter(IDs)

			result, err := client.GetHostDetailsRequest(ctx, IDs, hostFilter, orderBy, orderHow, limit, offset)

			// Assertions
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]*HostDetailsResponse{
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
			}))
		})

		It("Fails when no IDs are recieved", func() {
			client := NewXjoinClientMock()
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{}
			hostFilter := createHostsFilter(IDs)

			_, err := client.GetHostDetailsRequest(ctx, IDs, hostFilter, orderBy, orderHow, limit, offset)

			Expect(err).To(HaveOccurred())
		})

		It("Fails when the query returns an error", func() {
			client := NewXjoinClientMock()
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{}
			hostFilter := createHostsFilter(IDs)

			_, err := client.GetHostDetailsRequest(ctx, IDs, hostFilter, orderBy, orderHow, limit, offset)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetSystemProfileDetailsRequest", func() {
		It("Interperates response correctly", func() {
			client := NewXjoinClientMock()
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"1234", "5678"}
			hostFilter := createHostsFilter(IDs)

			result, err := client.GetSystemProfileDetailsRequest(ctx, IDs, hostFilter, orderBy, orderHow, limit, offset)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]*SystemProfileDetailsResponse{
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
			}))
		})

		It("Fails when no IDs are recieved", func() {
			client := NewXjoinClientMock()
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{}
			hostFilter := createHostsFilter(IDs)

			_, err := client.GetSystemProfileDetailsRequest(ctx, IDs, hostFilter, orderBy, orderHow, limit, offset)

			Expect(err).To(HaveOccurred())
		})

		It("Fails when the query returns an error", func() {
			client := NewXjoinClientMock()
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{}
			hostFilter := createHostsFilter(IDs)

			_, err := client.GetSystemProfileDetailsRequest(ctx, IDs, hostFilter, orderBy, orderHow, limit, offset)

			Expect(err).To(HaveOccurred())
		})
	})
})
