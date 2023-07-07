package private

import (
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/utils/test"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func sendJobRequest(payload *ApiInternalHighlevelJobRequestJSONRequestBody) *ApiInternalHighlevelJobRequestResponse {
	resp, err := client.ApiInternalHighlevelJobRequest(test.TestContext(), *payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalHighlevelJobRequestResponse(resp)
	Expect(err).ToNot(HaveOccurred())

	return res
}

var _ = Describe("high level job request", func() {
	It("successfully sends job request for a single recipient", func() {
		payload := minimalV2Payload(uuid.New())
		hostId1 := "c484f980-ab8d-401b-90e7-aa1d4ccf8c0e"
		payload.Hosts = &RunInputHosts{{InventoryId: &hostId1}}

		response := sendJobRequest(&ApiInternalHighlevelJobRequestJSONRequestBody{payload})
		result := response.JSON207

		Expect(response.StatusCode()).To(Equal(207))
		Expect(*result).To(HaveLen(1))
		Expect((*result)[0].Code).To(Equal(201))
		Expect((*result)[0].Id).ToNot(BeNil())
	})

	It("successfully sends job request for multiple recipients", func() {
		recipientInfo1 := minimalV2Payload(uuid.New())
		recipientInfo2 := minimalV2Payload(uuid.New())

		hostId1, hostId2 := "c484f980-ab8d-401b-90e7-aa1d4ccf8c0e", "0ecef95e-98b2-41aa-899a-b634c043b4d5"

		recipientInfo1.Hosts = &RunInputHosts{{InventoryId: &hostId1}}
		recipientInfo2.Hosts = &RunInputHosts{{InventoryId: &hostId2}}

		satId := "a24139f1-a28b-4b53-888d-d9e6ec5f8207"
		satOrgId := "1234"
		recpientConfig := RecipientConfig{SatId: &satId, SatOrgId: &satOrgId}
		recipientInfo2.RecipientConfig = &recpientConfig

		// recipient 1 is direct connected, recipient 2 is a satellite
		response := sendJobRequest(&ApiInternalHighlevelJobRequestJSONRequestBody{recipientInfo1, recipientInfo2})
		result := response.JSON207

		Expect(response.StatusCode()).To(Equal(207))
		Expect(*result).To(HaveLen(2))
		Expect((*result)[0].Code).To(Equal(201))
		Expect((*result)[0].Id).ToNot(BeNil())
		Expect((*result)[1].Code).To(Equal(201))
		Expect((*result)[1].Id).ToNot(BeNil())
	})

	DescribeTable("high level job request validation",
		func(payload, expected string) {
			resp, err := client.ApiInternalHighlevelJobRequestWithBody(test.TestContext(), "application/json", strings.NewReader(payload))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(ContainSubstring(expected))
		},

		Entry("empty list", `[]`, "Minimum number of items is 1"),

		Entry(
			"missing required property (org_id)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook"}]`,
			"Property 'org_id' is missing",
		),
		Entry(
			"missing required property (principal)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "url": "http://example.com", "name": "Red Hat Playbook"}]`,
			"Property 'principal' is missing",
		),
		Entry(
			"missing required property (name)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com"}]`,
			"Property 'name' is missing",
		),
		Entry(
			"missing required property (url)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "name": "Red Hat Playbook"}]`,
			"Property 'url' is missing",
		),
		Entry(
			"invalid property (org_id)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "53182905318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook"}]`,
			"Maximum string length is 10",
		),
		Entry(
			"timeout minimum",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "timeout": -1}]`,
			"Number must be at least 0",
		),
		Entry(
			"timeout maximum",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "timeout": 1000000}]`,
			"Number must be most 604800",
		),

		// min length
		Entry(
			"invalid property (principal)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "", "url": "http://example.com", "name": "Red Hat Playbook"}]`,
			"Minimum string length is 1",
		),
		Entry(
			"invalid property (name)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": ""}]`,
			"Minimum string length is 1",
		),
		Entry(
			"invalid property (web console url)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "web_console_url": ""}]`,
			"Minimum string length is 1",
		),
		Entry(
			"invalid property (ansible host)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "hosts": [{"ansible_host": ""}]}]`,
			"Minimum string length is 1",
		),

		Entry(
			"invalid URL",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "blahblah", "name": "Red Hat Playbook"}]`,
			`JSON string doesn't match the format \"url\"`,
		),
		Entry(
			"invalid URL (web console url)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "web_console_url": "blahblah"}]`,
			`JSON string doesn't match the format \"url\"`,
		),

		// invalid host data
		Entry(
			"host array not provided",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook"}]`,
			`Hosts need to be defined`,
		),
		Entry(
			"host array empty",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "hosts": []}]`,
			`Hosts cannot be empty`,
		),
		Entry(
			"inventory id not provided",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "hosts": [{"inventory_id": "591c0f11-67cb-4c64-bd6d-dc4ddb622f44"},{"ansible_host": "something"}]}]`,
			`Inventory ID missing in hosts array`,
		),

		// Satellite-specific
		Entry(
			"invalid Sattelite id",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "recipient_config": {"sat_id": "abc", "sat_org_id": "1"}}]`,
			`JSON string doesn't match the format \"sat-id-uuid\"`,
		),
		Entry(
			"invalid Sattelite org id",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "recipient_config": {"sat_id": "e7ee7fdd-b732-4eea-a070-40025b3dddd9", "sat_org_id": ""}}]`,
			`Minimum string length is 1`,
		),
		Entry(
			"sat_org_id missing",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "hosts": [{"inventory_id": "591c0f11-67cb-4c64-bd6d-dc4ddb622f44"}], "recipient_config": {"sat_id": "e7ee7fdd-b732-4eea-a070-40025b3dddd9"}}]`,
			`Both sat_id and sat_org need to be defined for satellite requests`,
		),
		Entry(
			"sat_org_id missing",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "hosts": [{"inventory_id": "591c0f11-67cb-4c64-bd6d-dc4ddb622f44"}], "recipient_config": {"sat_org_id": "1"}}]`,
			`Both sat_id and sat_org need to be defined for satellite requests`,
		),
	)
})
