package inventory

import (
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Inventory", func() {
	Describe("GetHostDetails", func() {
		It("Interperates response correctly", func() {
			doer := test.MockHttpClient(200, `{"results":[{"id":"1234","display_name":"test","facts":[{"namespace":"satellite", "facts":{"satellite_version": 6.11,"satellite_instance_id":"5678"}}],"fqdn":"test_host"}]}`)

			client := NewInventoryClientWithHttpRequestDoer(config.Get(), &doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"1234"}
			result, err := client.GetHostDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			resultData := result[0]
			Expect(err).ToNot(HaveOccurred())
			Expect(resultData.ID).To(Equal("1234"))
			Expect(resultData.DisplayName).To(Equal("test"))
			Expect(resultData.Facts).To(Equal(map[string]interface{}{
				"satellite_instance_id": "5678",
				"satellite_version":     6.11,
			},
			))
		})

		It("Interperates response correctly on unexpected status code", func() {
			doer := test.MockHttpClient(400, `{"results":[{"id":"1234","display_name":"test","facts":[{"namespace":"satellite", "facts":{"satellite_version": 6.11,"satellite_instance_id":"5678"}}],"fqdn":"test_host"}]}`)

			client := NewInventoryClientWithHttpRequestDoer(config.Get(), &doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"1234"}
			_, err := client.GetHostDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unexpected status code "400"`))
		})
	})

	Describe("GetSystemProfileDetails", func() {
		It("Interperates response correctly", func() {
			doer := test.MockHttpClient(200, `{"results":[{"id":"1234","system_profile":{"rhc_client_id":"7bc66a39-e719-4bc5-b10a-77bfbd3a0ead","owner_id":"b2ea37a0-7fb0-4f14-815d-fb582a916d5b"}}]}`)

			client := NewInventoryClientWithHttpRequestDoer(config.Get(), &doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"1234"}
			result, err := client.GetSystemProfileDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			resultData := result[0]
			Expect(err).ToNot(HaveOccurred())
			Expect(resultData.ID).To(Equal("1234"))
			Expect(resultData.SystemProfileFacts).To(Equal(map[string]interface{}{
				"rhc_client_id": "7bc66a39-e719-4bc5-b10a-77bfbd3a0ead",
				"owner_id":      "b2ea37a0-7fb0-4f14-815d-fb582a916d5b",
			},
			))
		})

		It("Interperates response correctly on unexpected status code", func() {
			doer := test.MockHttpClient(400, `{"results":[{"id":"1234","system_profile":{"rhc_client_id":"7bc66a39-e719-4bc5-b10a-77bfbd3a0ead","owner_id":"b2ea37a0-7fb0-4f14-815d-fb582a916d5b"}}]}`)

			client := NewInventoryClientWithHttpRequestDoer(config.Get(), &doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"1234"}
			_, err := client.GetSystemProfileDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unexpected status code "400"`))
		})
	})
})
