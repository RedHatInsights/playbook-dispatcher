package tenants

import (
	"playbook-dispatcher/internal/common/utils/test"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tenant Translator Suite")
}

var _ = Describe("Tenant Translator implementation", func() {
	Describe("lookup by RHC ID", func() {
		It("interprets the response correctly", func() {
			doer := test.MockHttpClient(200, `{"x-rh-identity": "eyJpZGVudGl0eSI6IHsib3JnX2lkIjogIjUzMTgyOTAiLCAiYWNjb3VudF9udW1iZXIiOiAiOTAxNTc4In19"}`)

			client := NewTenantIDTranslatorClient(
				"http://localhost",
				WithDoer(&doer),
			)

			orgID, ean, err := client.RHCIDToTenantIDs(test.TestContext(), "4112e0de-5237-441b-a4ac-e88713268f89")
			Expect(err).ToNot(HaveOccurred())
			Expect(orgID).To(Equal("5318290"))
			Expect(*ean).To(Equal("901578"))
		})

		It("handles tenant not found", func() {
			doer := test.MockHttpClient(401, `{}`)

			client := NewTenantIDTranslatorClient(
				"http://localhost",
				WithDoer(&doer),
			)

			_, _, err := client.RHCIDToTenantIDs(test.TestContext(), "4112e0de-5237-441b-a4ac-e88713268f89")
			Expect(err).To(HaveOccurred())
			_, ok := err.(*TenantNotFoundError)
			Expect(ok).To(BeTrue())
		})
	})
})
