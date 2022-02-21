package private

import (
	"net/http"
	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getStatus(payload ApiInternalV2RecipientsStatusJSONRequestBody) (*[]RecipientStatus, *ApiInternalV2RecipientsStatusResponse) {
	resp, err := client.ApiInternalV2RecipientsStatus(test.TestContext(), payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalV2RecipientsStatusResponse(resp)
	Expect(err).ToNot(HaveOccurred())
	Expect(res.StatusCode()).To(Equal(http.StatusOK))

	return res.JSON200, res
}

var _ = Describe("recipient status", func() {
	Describe("get status happy path", func() {
		It("single recipient", func() {

			payload := ApiInternalV2RecipientsStatusJSONRequestBody{
				RecipientWithOrg{
					OrgId:     "5318290",
					Recipient: "214f2dc3-eda5-4230-9800-579b020be25b",
				},
			}

			result, _ := getStatus(payload)

			Expect(*result).To(HaveLen(1))
			Expect((*result)[0].OrgId).To(Equal(payload[0].OrgId))
			Expect((*result)[0].Recipient).To(Equal(payload[0].Recipient))
			Expect((*result)[0].Connected).To(BeTrue())
		})

		It("multiple recipients", func() {

			payload := ApiInternalV2RecipientsStatusJSONRequestBody{
				RecipientWithOrg{
					OrgId:     "5318290",
					Recipient: "214f2dc3-eda5-4230-9800-579b020be25b",
				},
				RecipientWithOrg{
					OrgId:     "5318290",
					Recipient: "411cb203-f8c9-480e-ba20-1efbc74e3a33",
				},
			}

			result, _ := getStatus(payload)

			Expect(*result).To(HaveLen(2))
			Expect((*result)[0].OrgId).To(Equal(payload[0].OrgId))
			Expect((*result)[0].Recipient).To(Equal(payload[0].Recipient))
			Expect((*result)[0].Connected).To(BeTrue())
			Expect((*result)[1].OrgId).To(Equal(payload[1].OrgId))
			Expect((*result)[1].Recipient).To(Equal(payload[1].Recipient))
			Expect((*result)[1].Connected).To(BeFalse())
		})
	})

	It("400s on unknown tenant", func() {
		payload := ApiInternalV2RecipientsStatusJSONRequestBody{
			RecipientWithOrg{
				OrgId:     "654321",
				Recipient: "ac2edabb-c489-4d28-8c7b-d60fbd2f99ba",
			},
		}

		resp, err := client.ApiInternalV2RecipientsStatus(test.TestContext(), payload)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(400))
	})

	It("400s on anemic tenant", func() {
		payload := ApiInternalV2RecipientsStatusJSONRequestBody{
			RecipientWithOrg{
				OrgId:     "654322",
				Recipient: "ac2edabb-c489-4d28-8c7b-d60fbd2f99ba",
			},
		}

		resp, err := client.ApiInternalV2RecipientsStatus(test.TestContext(), payload)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(400))
	})
})
