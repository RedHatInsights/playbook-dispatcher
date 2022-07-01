package private

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Private Suite")
}

var _ = Describe("Validation", func() {
	DescribeTable("validateFields",
		func(runInputJson string, valid bool) {
			runInput := RunInputV2{}
			byteString := []byte(runInputJson)

			err := json.Unmarshal(byteString, &runInput)
			Expect(err).To(BeNil())

			err = validateSatelliteFields(runInput)

			Expect(err == nil).To(Equal(valid))
		},

		Entry(
			"valid ansible",
			`{"org_id": "123", "recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c", "url": "http://example.com", "principal": "test-user","hosts": [{"ansible_host": "localhost","inventory_id": "16372e6f-1c18-4cdb-b780-50ab4b88e74b"}]}`,
			true,
		),
		Entry("valid satellite",
			`{"org_id": "123","recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c","url": "http://example.com","playbook_name":"test-playbook","playbook_run_url": "http://example.com","recipient_config": {"sat_id":"16372e6f-1c18-4cdb-b780-50ab4b88e74b","sat_org_id":"456"},"principal": "test-user","hosts": [{"inventory_id": "16372e6f-1c18-4cdb-b780-50ab4b88e74b"}, {"inventory_id": "baf2bb2f-06a3-42cc-ae7b-68ccc8e2a344"}]}`,
			true,
		),
		Entry("satellite - missing sat_id",
			`{"org_id": "123","recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c","url": "http://example.com","playbook_name":"test-playbook","playbook_run_url": "http://example.com","recipient_config": {"sat_org_id":"456"},"principal": "test-user","hosts": [{"inventory_id": "16372e6f-1c18-4cdb-b780-50ab4b88e74b"}, {"inventory_id": "baf2bb2f-06a3-42cc-ae7b-68ccc8e2a344"}]}`,
			false,
		),
		Entry("satellite - missing sat_org_id",
			`{"org_id": "123","recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c","url": "http://example.com","playbook_name":"test-playbook","playbook_run_url": "http://example.com","recipient_config": {"sat_id":"16372e6f-1c18-4cdb-b780-50ab4b88e74b"},"principal": "test-user","hosts": [{"inventory_id": "16372e6f-1c18-4cdb-b780-50ab4b88e74b"}, {"inventory_id": "baf2bb2f-06a3-42cc-ae7b-68ccc8e2a344"}]}`,
			false,
		),
		Entry("satellite - hosts missing",
			`{"org_id": "123","recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c","url": "http://example.com","playbook_name":"test-playbook","playbook_run_url": "http://example.com","recipient_config": {"sat_id":"16372e6f-1c18-4cdb-b780-50ab4b88e74b","sat_org_id":"456"},"principal": "test-user"}`,
			false,
		),
		Entry("satellite - hosts empty",
			`{"org_id": "123","recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c","url": "http://example.com","playbook_name":"test-playbook","playbook_run_url": "http://example.com","recipient_config": {"sat_id":"16372e6f-1c18-4cdb-b780-50ab4b88e74b","sat_org_id":"456"},"principal": "test-user", "hosts": []}`,
			false,
		),
	)
})
