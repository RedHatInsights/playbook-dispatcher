package validator

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils/test"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/qri-io/jsonschema"
)

var instance handler

var _ = Describe("Handler", func() {
	BeforeEach(func() {
		var schemas []*jsonschema.Schema

		for _, filePath := range []string{"../../schema/ansibleRunnerJobEvent.yaml", "../../schema/rhcsatJobEvent.yaml"} {
			var schema jsonschema.Schema
			file, err := ioutil.ReadFile(filePath)
			Expect(err).ToNot(HaveOccurred())
			err = yaml.Unmarshal(file, &schema)
			Expect(err).ToNot(HaveOccurred())

			schemas = append(schemas, &schema)
		}

		instance = handler{
			producer: nil,
			schemas:  schemas,
		}
	})

	Describe("File size", func() {
		It("Rejects an upload over max size", func() {
			req := &messageModel.IngressValidationRequest{
				Size: 128 * 1024 * 1024,
			}

			err := instance.validateRequest(req)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Validation", func() {

		DescribeTable("Rejects invalid files",
			func(requestType string, file string) {
				_, err := instance.validateContent(test.TestContext(), requestType, []byte(file))
				Expect(err).To(HaveOccurred())
			},
			Entry("empty file", "playbook", ""),
			Entry("whitespace-only file", "playbook", "         "),
			Entry("newline-only file", "playbook", "\n\n\n\n"),
			Entry("invalid JSON (trailing braces)", "playbook", `{"event": "playbook_on_start", "uuid": "60049c81-4f4b-41a0-bcf4-84399bf1b693", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}}`),
			Entry("missing uuid", "playbook", `{"event": "playbook_on_start", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}`),

			Entry("rhc-sat newline-only file", "playbook-sat", "\n\n\n\n"),
			Entry("missing rhc-sat uuid", "playbook-sat", `{"type" : "playbook_run_update", "version": 3}`),
			Entry("rhc-sat invalid JSON", "playbook-sat", `{"type" : "playbook_run_update", "version": 3, "correlation_id": "0465783c-2e36-4e57-8514-c2cb962d323a"}}`),
			Entry("rhc-sat invalid type", "playbook-sat", `{"type" : "invalid_type", "version": 3, "correlation_id": "0465783c-2e36-4e57-8514-c2cb962d323a"}`),
		)

		DescribeTable("Accepts valid runner files",
			func(requestType string, file string) {
				events, err := instance.validateContent(test.TestContext(), requestType, []byte(file))
				Expect(err).ToNot(HaveOccurred())
				Expect(events.Playbook).ToNot(BeEmpty())
			},

			Entry("multiple events", "playbook", `
			{"event": "playbook_on_start", "uuid": "cb93301e-5ff8-4f75-ade6-57d0ec2fc662", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}
			{"event": "playbook_on_stats", "uuid": "998a4bd2-2d6b-4c31-905c-2d5ad7a7f8ab", "counter": 1, "stdout": "", "start_line": 0, "end_line": 0}
			`),

			Entry("extra attributes", "playbook", `{"event": "playbook_on_start", "uuid": "cb93301e-5ff8-4f75-ade6-57d0ec2fc662", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0, "event_data": {"playbook": "ping.yml", "playbook_uuid": "db6da5c7-37a6-479f-b18a-1db5af7f0932", "uuid": "db6da5c7-37a6-479f-b18a-1db5af7f0932"}}`),
		)

		DescribeTable("Accepts valid rhc-sat files",
			func(requestType string, file string) {
				events, err := instance.validateContent(test.TestContext(), requestType, []byte(file))
				Expect(err).ToNot(HaveOccurred())
				Expect(events.PlaybookSat).To(HaveLen(3))
				Expect(events.Playbook).To(BeEmpty())
			},

			Entry("multiple events", "playbook-sat", `
			{"type": "playbook_run_update", "version": 3, "correlation_id": "0465783c-2e36-4e57-8514-c2cb962d323a", "sequence": 1, "host": "03.example.com", "console": "03.example.com | SUCCESS => {\n    \"changed\": false,\n    \"ping\": \"pong\"\n}"}
			{"type": "playbook_run_finished", "version": 3, "correlation_id": "0465783c-2e36-4e57-8514-c2cb962d323a", "host": "03.example.com", "status": "success", "connection_code": 0, "execution_code": 0}
			{"type": "playbook_run_completed", "version": 3, "correlation_id": "0465783c-2e36-4e57-8514-c2cb962d323a", "status": "success", "satellite_connection_code": 0, "satellite_infrastructure_code": 0}
			`),
		)

		It("parses Runner events", func() {
			data := `
			{"uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "counter": 1, "stdout": null, "start_line": 0, "end_line": 0, "runner_ident": "test05", "event": "playbook_on_start", "pid": 1149259, "created": "2021-01-22T14:41:59.728652", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713"}}
			{"uuid": "58961d98-604d-ab6c-a789-000000000008", "counter": 2, "stdout": "\r\nPLAY [ping] ********************************************************************", "start_line": 0, "end_line": 2, "runner_ident": "test05", "event": "playbook_on_play_start", "pid": 1149259, "created": "2021-01-22T14:41:59.730167", "parent_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "name": "ping", "pattern": "localhost", "uuid": "58961d98-604d-ab6c-a789-000000000008"}}
			{"uuid": "58961d98-604d-ab6c-a789-00000000000a", "counter": 3, "stdout": "\r\nTASK [ping] ********************************************************************", "start_line": 2, "end_line": 4, "runner_ident": "test05", "event": "playbook_on_task_start", "pid": 1149259, "created": "2021-01-22T14:41:59.735735", "parent_uuid": "58961d98-604d-ab6c-a789-000000000008", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "task": "ping", "task_uuid": "58961d98-604d-ab6c-a789-00000000000a", "task_action": "ping", "task_args": "", "task_path": "minimal.yml:5", "name": "ping", "is_conditional": false, "uuid": "58961d98-604d-ab6c-a789-00000000000a"}}
			{"uuid": "bdca6550-db72-44bc-a0e2-a1f2dc25f3e5", "counter": 4, "stdout": "", "start_line": 4, "end_line": 4, "runner_ident": "test05", "event": "runner_on_start", "pid": 1149259, "created": "2021-01-22T14:41:59.736129", "parent_uuid": "58961d98-604d-ab6c-a789-00000000000a", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "task": "ping", "task_uuid": "58961d98-604d-ab6c-a789-00000000000a", "task_action": "ping", "task_args": "", "task_path": "minimal.yml:5", "host": "localhost", "uuid": "bdca6550-db72-44bc-a0e2-a1f2dc25f3e5"}}
			{"uuid": "7ea01103-2b54-4128-a838-52d0113e455d", "counter": 5, "stdout": "\u001b[0;32mok: [localhost]\u001b[0m", "start_line": 4, "end_line": 5, "runner_ident": "test05", "event": "runner_on_ok", "pid": 1149259, "created": "2021-01-22T14:42:00.006719", "parent_uuid": "58961d98-604d-ab6c-a789-00000000000a", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "task": "ping", "task_uuid": "58961d98-604d-ab6c-a789-00000000000a", "task_action": "ping", "task_args": "", "task_path": "minimal.yml:5", "host": "localhost", "remote_addr": "127.0.0.1", "res": {"ping": "pong", "invocation": {"module_args": {"data": "pong"}}, "_ansible_no_log": false, "changed": false}, "event_loop": null, "uuid": "7ea01103-2b54-4128-a838-52d0113e455d"}}
			{"uuid": "c8347ac2-61d3-4a36-9cbb-c51e14984eee", "counter": 6, "stdout": "\r\nPLAY RECAP *********************************************************************\r\n\u001b[0;32mlocalhost\u001b[0m                  : \u001b[0;32mok=1   \u001b[0m changed=0    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   \r\n", "start_line": 5, "end_line": 9, "runner_ident": "test05", "event": "playbook_on_stats", "pid": 1149259, "created": "2021-01-22T14:42:00.009228", "parent_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "changed": {}, "dark": {}, "failures": {}, "ignored": {}, "ok": {"localhost": 1}, "processed": {"localhost": 1}, "rescued": {}, "skipped": {}, "artifact_data": {}, "uuid": "c8347ac2-61d3-4a36-9cbb-c51e14984eee"}}
			`

			events, err := instance.validateContent(test.TestContext(), "playbook", []byte(data))
			Expect(err).ToNot(HaveOccurred())
			Expect(events.Playbook).To(HaveLen(6))
		})
	})

	Describe("compression", func() {
		It("parses compressed Runner events", func() {
			data := `H4sICMI0KmAAA2Zvby5qc29ubADtV8tu2zAQvPcrDB2D0iApUg8XOQRFT+0haHMpkkCgSMoRLJGG
HgECw/9ekpYUq06ayEmLoIjgg3e5S652ZlbSxmvbXHiLmScIkzHlGQhRJgDxowBEIs0ATrmMfZ/L
EPnex5nHdasaWZkUZKy6EbptbL7nLFY1SZEraTzQOKQS+2bVKiWrJBdSuZxG1g2kNlPedq51we5S
rVeJVonbzq6uXYkIkRjT2NZQSdZIVzaGGAGIAMYXiCwIWtB4HuIooHjYNhGsYSZ2M2xuE8tc5SUr
5ndl4Y7oz53aj0nx2+2HzZBBozhAIo5AAIkALA04YGEUA3h/ReOO41HHr6ordf7t7Ofscp2r5fXs
5BWup1DE01G0/4+H0ocoCF0iqyyWU+H5Bwywie6mDQq9nUwF2SWtWWOAVjav0JwVN7p2TVOslKMT
HombdOhULrIxF/0DLl6c/fj6V7mIx1wk07nYsHr1Ai5S83uAi8/F+L/gom3h/gmupVNp5JIYb3Kt
DvZi1bLuHyjOYSq5+a1HC/qQLPI64VqJ3G7LCrOSsaKW03TBxrpIBWcBpRCINMSAkNRkQIkBQxkW
HNPMl3SsC/LnpyI5isNdyEueiX6AcHwcedk7eV+fvK6kxwb4s2g3ImooGUQI+qaplACCcARY5EeA
YmH8viSUijFR6XiAtxCi9BJ+8nGpV4vZ5VDXdb9UPkVmOpXMhj6TmIwXEM4hDEL0zuQ3zuRKlrqR
CRPCcs1DOJwb6OZot1bvOm/Ps+fqbn6rW7NFV83GK7VoC9kXsvE6xHbh261JSJiq89TEKJ0Uerk3
8fkNU0vHIefZDrAXWq+NV7VFsae3Z6lnpDce+SRkHAPTaB8Q5gcg5mkKOEXSMDkiUsqx3oKHX96/
f/l8dv4q70snds99HQ+ADAqeHVyL2Vj5p8g47+O7Pp5CG9sqo0zjMR3fOTKWF/1ivcrX694wCPO2
N/Kl0lVn2BIPxggdj5H4qC/Fpj5qlsQYR2/78+KeyhtLY8GqVf/f9r/t1GTtrtO96erY7OnSdMb6
15Xmsq53gYfLHXj9Lh2wvWlwyzMzM4ab3u7p6FmqMDr6BVcdZ0B3EAAA`

			decoded := make([]byte, 1024)
			len, err := base64.StdEncoding.Decode(decoded, []byte(data))
			Expect(err).ToNot(HaveOccurred())

			content, err := readFile(bytes.NewReader(decoded[0:len]))
			Expect(err).ToNot(HaveOccurred())
			events, err := instance.validateContent(test.TestContext(), "playbook", content)
			Expect(err).ToNot(HaveOccurred())
			Expect(events.Playbook).To(HaveLen(6))
		})
	})

	// TODO: test parsing (timestamps, etc.)
})
