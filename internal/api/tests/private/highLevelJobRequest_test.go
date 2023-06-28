package private

import (
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func sendJobRequest(payload ApiInternalHighlevelJobRequestJSONRequestBody) *ApiInternalHighlevelJobRequestResponse {
	resp, err := client.ApiInternalHighlevelJobRequest(test.TestContext(), payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalHighlevelJobRequestResponse(resp)
	Expect(err).ToNot(HaveOccurred())

	return res
}

var _ = Describe("high level job request", func() {
	It("successfully sends job request to multiple hosts", func() {
		satelliteHost := []HostId{"c484f980-ab8d-401b-90e7-aa1d4ccf8c0e"}
		directConnectHost := []HostId{"fe30b997-c15a-44a9-89df-c236c3b5c540"}

		payload := ApiInternalHighlevelJobRequestJSONRequestBody{
			Hosts:        []string{"c484f980-ab8d-401b-90e7-aa1d4ccf8c0e", "fe30b997-c15a-44a9-89df-c236c3b5c540"},
			OrgId:        "12345",
			Principal:    "testuser",
			Url:          "https://example.com",
			PlaybookName: "testPlaybook",
		}

		response := sendJobRequest(payload)
		result := response.JSON200

		Expect(response.StatusCode()).To(Equal(200))
		Expect(*result).To(HaveLen(2))
		Expect((*result)[0].Recipient).To(Equal(public.RunRecipient("d415fc2d-9700-4e30-9621-6a410ccc92d8")))
		Expect((*result)[0].RequestDispatch).To(Equal("success"))
		Expect((*result)[0].Status).To(Equal("connected"))
		Expect((*result)[0].Systems).To(Equal(satelliteHost))
		Expect((*result)[0].RunId).ToNot(BeNil())

		Expect((*result)[1].Recipient).To(Equal(public.RunRecipient("32af5948-301f-449a-a25b-ff34c83264a2")))
		Expect((*result)[1].RequestDispatch).To(Equal("success"))
		Expect((*result)[1].Status).To(Equal("connected"))
		Expect((*result)[1].Systems).To(Equal(directConnectHost))
		Expect((*result)[1].RunId).ToNot(BeNil())
	})

	It("sends 400 on bad request body", func() {
		// No principal provided in request body
		payload := ApiInternalHighlevelJobRequestJSONRequestBody{
			Hosts:        []string{"c484f980-ab8d-401b-90e7-aa1d4ccf8c0e"},
			OrgId:        "12345",
			Url:          "https://example.com",
			PlaybookName: "testPlaybook",
		}

		response := sendJobRequest(payload)
		Expect(response.StatusCode()).To(Equal(400))

		result := response.JSON400
		Expect(result.Message).To(Equal("Request body has an error: doesn't match the schema: Error at \"/principal\": Minimum string length is 1"))
	})
})
