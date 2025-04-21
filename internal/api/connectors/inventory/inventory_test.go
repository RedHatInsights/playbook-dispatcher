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
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"results":[{"id":"1234","display_name":"test","facts":[{"namespace":"satellite", "facts":{"satellite_version": "6.11.3","satellite_instance_id":"5678","organization_id":"5"}}],"fqdn":"test_host"}]}`},
				{StatusCode: 200, Body: `{"results":[{"id":"1234","system_profile":{"rhc_client_id":"7bc66a39-e719-4bc5-b10a-77bfbd3a0ead","owner_id":"b2ea37a0-7fb0-4f14-815d-fb582a916d5b"}}]}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewInventoryClientWithHttpRequestDoer(config.Get(), doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"db0b6f08-e0ba-4248-8e0e-2de2fb843dcf"}
			result, err := client.GetHostConnectionDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			resultData := result[0]
			Expect(err).ToNot(HaveOccurred())
			Expect(resultData.ID).To(Equal("1234"))
			Expect(*resultData.OwnerID).To(Equal("b2ea37a0-7fb0-4f14-815d-fb582a916d5b"))
			Expect(*resultData.SatelliteInstanceID).To(Equal("5678"))
			Expect(*resultData.SatelliteVersion).To(Equal("6.11.3"))
			Expect(*resultData.SatelliteOrgID).To(Equal("5"))
			Expect(*resultData.RHCClientID).To(Equal("7bc66a39-e719-4bc5-b10a-77bfbd3a0ead"))
		})

		It("Interperates response correctly on unexpected status code from hostDetails", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 400, Body: `{"results":[{"id":"1234","display_name":"test","facts":[{"namespace":"satellite", "facts":{"satellite_version": "6.11.3","satellite_instance_id":"5678"}}],"fqdn":"test_host"}]}`},
				{StatusCode: 200, Body: `{"results":[{"id":"1234","system_profile":{"rhc_client_id":"7bc66a39-e719-4bc5-b10a-77bfbd3a0ead","owner_id":"b2ea37a0-7fb0-4f14-815d-fb582a916d5b"}}]}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewInventoryClientWithHttpRequestDoer(config.Get(), doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"db0b6f08-e0ba-4248-8e0e-2de2fb843dcf"}
			_, err := client.GetHostConnectionDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unexpected status code "400"`))
		})

		It("Interperates response correctly on unexpected status code from systemProfileDetails", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"results":[{"id":"1234","display_name":"test","facts":[{"namespace":"satellite", "facts":{"satellite_version": "6.11.3","satellite_instance_id":"5678"}}],"fqdn":"test_host"}]}`},
				{StatusCode: 400, Body: `{"results":[{"id":"1234","system_profile":{"rhc_client_id":"7bc66a39-e719-4bc5-b10a-77bfbd3a0ead","owner_id":"b2ea37a0-7fb0-4f14-815d-fb582a916d5b"}}]}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewInventoryClientWithHttpRequestDoer(config.Get(), doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			IDs := []string{"db0b6f08-e0ba-4248-8e0e-2de2fb843dcf"}
			_, err := client.GetHostConnectionDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unexpected status code "400"`))
		})

		It("Interperates response correctly when host detail facts are not present", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"results":[{"id":"1234","display_name":"test","facts":[],"fqdn":"test_host"}]}`},
				{StatusCode: 200, Body: `{"results":[{"id":"1234","system_profile":{"rhc_client_id":"7bc66a39-e719-4bc5-b10a-77bfbd3a0ead"}}]}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewInventoryClientWithHttpRequestDoer(config.Get(), doer)
			ctx := test.TestContext()
			IDs := []string{"db0b6f08-e0ba-4248-8e0e-2de2fb843dcf"}
			result, err := client.GetHostConnectionDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			resultData := result[0]
			Expect(err).ToNot(HaveOccurred())
			Expect(resultData.ID).To(Equal("1234"))
			Expect(resultData.OwnerID).To(BeNil())
			Expect(resultData.SatelliteInstanceID).To(BeNil())
			Expect(resultData.SatelliteVersion).To(BeNil())
			Expect(resultData.SatelliteOrgID).To(BeNil())
			Expect(*resultData.RHCClientID).To(Equal("7bc66a39-e719-4bc5-b10a-77bfbd3a0ead"))
		})

		It("Interperates response correctly when rhc_client_id is not present", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"results":[{"id":"1234","display_name":"test","facts":[{"namespace":"satellite", "facts":{"satellite_version": "6.11.3","satellite_instance_id":"5678","organization_id":"5"}}],"fqdn":"test_host"}]}`},
				{StatusCode: 200, Body: `{"results":[{"id":"1234","system_profile":{"owner_id":"b2ea37a0-7fb0-4f14-815d-fb582a916d5b"}}]}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewInventoryClientWithHttpRequestDoer(config.Get(), doer)
			ctx := test.TestContext()
			IDs := []string{"db0b6f08-e0ba-4248-8e0e-2de2fb843dcf"}
			result, err := client.GetHostConnectionDetails(ctx, IDs, "DisplayName", "ASC", 10, 0)
			resultData := result[0]
			Expect(err).ToNot(HaveOccurred())
			Expect(resultData.ID).To(Equal("1234"))
			Expect(*resultData.OwnerID).To(Equal("b2ea37a0-7fb0-4f14-815d-fb582a916d5b"))
			Expect(*resultData.SatelliteInstanceID).To(Equal("5678"))
			Expect(*resultData.SatelliteVersion).To(Equal("6.11.3"))
			Expect(*resultData.SatelliteOrgID).To(Equal("5"))
			Expect(resultData.RHCClientID).To(BeNil())
		})
	})
})
