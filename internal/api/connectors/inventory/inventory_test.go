package inventory

import (
	"fmt"
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
			fmt.Print("RESULT: ", result)
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
	})

	// Describe("GetSystemProfileDetails", func() {
	// 	It("Interperates response correctly", func() {
	// 		// Add Test here
	// 	})
	// })
})
