package validator

import (
	"io/ioutil"
	messageModel "playbook-dispatcher/internal/common/model/message"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/qri-io/jsonschema"
	"go.uber.org/zap"
)

var instance handler

var _ = Describe("Handler", func() {
	BeforeEach(func() {
		var schema jsonschema.Schema
		file, err := ioutil.ReadFile("../../schema/ansibleRunnerJobEvent.yaml")
		Expect(err).ToNot(HaveOccurred())
		err = yaml.Unmarshal(file, &schema)
		Expect(err).ToNot(HaveOccurred())

		log := zap.NewNop().Sugar()
		instance = handler{
			log:      log,
			producer: nil,
			schema:   &schema,
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
			func(file string) {
				_, err := instance.validateContent([]byte(file))
				Expect(err).To(HaveOccurred())
			},
			Entry("empty file", ""),
			Entry("whitespace-only file", "         "),
			Entry("newline-only file", "\n\n\n\n"),
			Entry("invalid JSON (trailing braces)", `{"event": "playbook_on_start", "uuid": "60049c81-4f4b-41a0-bcf4-84399bf1b693", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}}`),
			Entry("missing uuid", `{"event": "playbook_on_start", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}`),
		)

		DescribeTable("Accepts valid files",
			func(file string) {
				events, err := instance.validateContent([]byte(file))
				Expect(err).ToNot(HaveOccurred())
				Expect(events).ToNot(BeEmpty())
			},

			Entry("multiple events", `
			{"event": "playbook_on_start", "uuid": "cb93301e-5ff8-4f75-ade6-57d0ec2fc662", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}
			{"event": "playbook_on_stats", "uuid": "998a4bd2-2d6b-4c31-905c-2d5ad7a7f8ab", "counter": 1, "stdout": "", "start_line": 0, "end_line": 0}
			`),

			Entry("extra attributes", `{"event": "playbook_on_start", "uuid": "cb93301e-5ff8-4f75-ade6-57d0ec2fc662", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0, "event_data": {"playbook": "ping.yml", "playbook_uuid": "db6da5c7-37a6-479f-b18a-1db5af7f0932", "uuid": "db6da5c7-37a6-479f-b18a-1db5af7f0932"}}`),
		)

		It("parses Runner events", func() {
			data := `
			{"uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "counter": 1, "stdout": "", "start_line": 0, "end_line": 0, "runner_ident": "test05", "event": "playbook_on_start", "pid": 1149259, "created": "2021-01-22T14:41:59.728652", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713"}}
			{"uuid": "58961d98-604d-ab6c-a789-000000000008", "counter": 2, "stdout": "\r\nPLAY [ping] ********************************************************************", "start_line": 0, "end_line": 2, "runner_ident": "test05", "event": "playbook_on_play_start", "pid": 1149259, "created": "2021-01-22T14:41:59.730167", "parent_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "name": "ping", "pattern": "localhost", "uuid": "58961d98-604d-ab6c-a789-000000000008"}}
			{"uuid": "58961d98-604d-ab6c-a789-00000000000a", "counter": 3, "stdout": "\r\nTASK [ping] ********************************************************************", "start_line": 2, "end_line": 4, "runner_ident": "test05", "event": "playbook_on_task_start", "pid": 1149259, "created": "2021-01-22T14:41:59.735735", "parent_uuid": "58961d98-604d-ab6c-a789-000000000008", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "task": "ping", "task_uuid": "58961d98-604d-ab6c-a789-00000000000a", "task_action": "ping", "task_args": "", "task_path": "minimal.yml:5", "name": "ping", "is_conditional": false, "uuid": "58961d98-604d-ab6c-a789-00000000000a"}}
			{"uuid": "bdca6550-db72-44bc-a0e2-a1f2dc25f3e5", "counter": 4, "stdout": "", "start_line": 4, "end_line": 4, "runner_ident": "test05", "event": "runner_on_start", "pid": 1149259, "created": "2021-01-22T14:41:59.736129", "parent_uuid": "58961d98-604d-ab6c-a789-00000000000a", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "task": "ping", "task_uuid": "58961d98-604d-ab6c-a789-00000000000a", "task_action": "ping", "task_args": "", "task_path": "minimal.yml:5", "host": "localhost", "uuid": "bdca6550-db72-44bc-a0e2-a1f2dc25f3e5"}}
			{"uuid": "7ea01103-2b54-4128-a838-52d0113e455d", "counter": 5, "stdout": "\u001b[0;32mok: [localhost]\u001b[0m", "start_line": 4, "end_line": 5, "runner_ident": "test05", "event": "runner_on_ok", "pid": 1149259, "created": "2021-01-22T14:42:00.006719", "parent_uuid": "58961d98-604d-ab6c-a789-00000000000a", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "play": "ping", "play_uuid": "58961d98-604d-ab6c-a789-000000000008", "play_pattern": "localhost", "task": "ping", "task_uuid": "58961d98-604d-ab6c-a789-00000000000a", "task_action": "ping", "task_args": "", "task_path": "minimal.yml:5", "host": "localhost", "remote_addr": "127.0.0.1", "res": {"ping": "pong", "invocation": {"module_args": {"data": "pong"}}, "_ansible_no_log": false, "changed": false}, "event_loop": null, "uuid": "7ea01103-2b54-4128-a838-52d0113e455d"}}
			{"uuid": "c8347ac2-61d3-4a36-9cbb-c51e14984eee", "counter": 6, "stdout": "\r\nPLAY RECAP *********************************************************************\r\n\u001b[0;32mlocalhost\u001b[0m                  : \u001b[0;32mok=1   \u001b[0m changed=0    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   \r\n", "start_line": 5, "end_line": 9, "runner_ident": "test05", "event": "playbook_on_stats", "pid": 1149259, "created": "2021-01-22T14:42:00.009228", "parent_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "event_data": {"playbook": "minimal.yml", "playbook_uuid": "d4ae95cf-71fd-4386-8dbf-2bce933ce713", "changed": {}, "dark": {}, "failures": {}, "ignored": {}, "ok": {"localhost": 1}, "processed": {"localhost": 1}, "rescued": {}, "skipped": {}, "artifact_data": {}, "uuid": "c8347ac2-61d3-4a36-9cbb-c51e14984eee"}}
			`

			events, err := instance.validateContent([]byte(data))
			Expect(err).ToNot(HaveOccurred())
			Expect(events).To(HaveLen(6))
		})
	})

	// TODO: test parsing (timestamps, etc.)
})
