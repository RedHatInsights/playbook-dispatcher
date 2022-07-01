package sources

import (
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sources", func() {
	Describe("GetSourceConnectionDetails", func() {
		It("interperates response correctly", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"data": [{"id": "1", "name": "test", "availability_status": "connected"}]}`},
				{StatusCode: 200, Body: `{"id": "1", "rhc_id": "6f37c752ba1c48b1bcf74ef8f585d8ee", "availability_status": "connected"}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewSourcesClientWithHttpRequestDoer(config.Get(), doer)
			ctx := test.TestContext()

			rhcId := "6f37c752ba1c48b1bcf74ef8f585d8ee"
			availabilityStatus := "connected"
			sourceName := "test"

			result, err := client.GetSourceConnectionDetails(ctx, "4f37c752-ba1c-48b1-bcf7-4ef8f585d9ee")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(SourceConnectionStatus{
				ID:                 "1",
				SourceName:         &sourceName,
				RhcID:              &rhcId,
				AvailabilityStatus: &availabilityStatus,
			}))
		})

		It("interperates response correctly if fields are missing", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"data": [{"id": "1", "name": "test"}]}`},
				{StatusCode: 200, Body: `{"id": "1"}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewSourcesClientWithHttpRequestDoer(config.Get(), doer)
			ctx := test.TestContext()

			sourceName := "test"

			result, err := client.GetSourceConnectionDetails(ctx, "4f37c752-ba1c-48b1-bcf7-4ef8f585d9ee")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(SourceConnectionStatus{
				ID:                 "1",
				SourceName:         &sourceName,
				RhcID:              nil,
				AvailabilityStatus: nil,
			}))
		})

		It("interperates response correctly if getSources returns a 400", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 400, Body: `{}`},
				{StatusCode: 200, Body: `{"id": "1", "rhc_id": "6f37c752ba1c48b1bcf74ef8f585d8ee", "availability_status": "connected"}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewSourcesClientWithHttpRequestDoer(config.Get(), doer)
			ctx := test.TestContext()

			_, err := client.GetSourceConnectionDetails(ctx, "4f37c752-ba1c-48b1-bcf7-4ef8f585d9ee")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Source Bad Request"))
		})

		It("interperates response correctly if getRhcConnectionStatus returns a 404", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"data": [{"id": "1", "name": "test", "availability_status": "connected"}]}`},
				{StatusCode: 404, Body: `{}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewSourcesClientWithHttpRequestDoer(config.Get(), doer)
			ctx := test.TestContext()

			_, err := client.GetSourceConnectionDetails(ctx, "4f37c752-ba1c-48b1-bcf7-4ef8f585d9ee")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RHCStatus Not Found"))
		})

		It("interperates response correctly if getRhcConnectionStatus returns a 400", func() {
			responses := []test.MockHttpResponse{
				{StatusCode: 200, Body: `{"data": [{"id": "1", "name": "test", "availability_status": "connected"}]}`},
				{StatusCode: 400, Body: `{"id": "1", "rhc_id": "6f37c752ba1c48b1bcf74ef8f585d8ee", "availability_status": "connected"}`},
			}

			doer := test.MockMultiResponseHttpClient(responses...)
			client := NewSourcesClientWithHttpRequestDoer(config.Get(), doer)
			ctx := test.TestContext()

			_, err := client.GetSourceConnectionDetails(ctx, "4f37c752-ba1c-48b1-bcf7-4ef8f585d9ee")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RHCStatus Bad Request"))
		})
	})
})
