package ansible

import (
	"encoding/json"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func loadFile(path string) (events []messageModel.PlaybookRunResponseMessageYamlEventsElem) {
	file, err := os.ReadFile(path)
	Expect(err).ToNot(HaveOccurred())

	lines := strings.Split(string(file), "\n")

	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		event := messageModel.PlaybookRunResponseMessageYamlEventsElem{}
		err = json.Unmarshal([]byte(line), &event)
		Expect(err).ToNot(HaveOccurred())

		events = append(events, event)
	}

	return events
}

var _ = Describe("Ansible", func() {
	Describe("host", func() {
		It("determines ansible host from a successful run", func() {
			events := loadFile("./test-events1.jsonl")
			hosts := GetAnsibleHosts(events)
			Expect(hosts).To(HaveLen(1))
			Expect(hosts[0]).To(Equal("localhost"))
		})

		It("determines ansible host from a failed run", func() {
			events := loadFile("./test-events2.jsonl")
			hosts := GetAnsibleHosts(events)
			Expect(hosts).To(HaveLen(1))
			Expect(hosts[0]).To(Equal("localhost"))
		})

		It("determines ansible host from an incomplete run", func() {
			events := loadFile("./test-events3.jsonl")
			hosts := GetAnsibleHosts(events)
			Expect(hosts).To(HaveLen(1))
			Expect(hosts[0]).To(Equal("localhost"))
		})

		It("determines ansible host from an incomplete run (2)", func() {
			events := loadFile("./test-events4.jsonl")
			hosts := GetAnsibleHosts(events)
			Expect(hosts).To(HaveLen(0))
		})

		It("determines ansible hosts from a multi-host run", func() {
			events := loadFile("./test-events5.jsonl")
			hosts := GetAnsibleHosts(events)
			Expect(hosts).To(HaveLen(2))
			Expect(hosts[0]).To(Equal("jharting1"))
			Expect(hosts[1]).To(Equal("jharting2"))
		})
	})

	Describe("stdout", func() {
		It("determines stdout from a successful run", func() {
			events := loadFile("./test-events1.jsonl")
			stdout := GetStdout(events, nil)
			Expect(stdout).To(Equal("\r\nPLAY [ping] ********************************************************************\n\r\nTASK [ping] ********************************************************************\n\x1b[0;32mok: [localhost]\x1b[0m\n\r\nPLAY RECAP *********************************************************************\r\n\x1b[0;32mlocalhost\x1b[0m                  : \x1b[0;32mok=1   \x1b[0m changed=0    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   \r\n\n"))
		})

		It("determines stdout from a failed run", func() {
			events := loadFile("./test-events2.jsonl")
			stdout := GetStdout(events, nil)
			Expect(stdout).To(Equal("\r\nPLAY [ping] ********************************************************************\n\r\nTASK [fail] ********************************************************************\n\x1b[0;31mfatal: [localhost]: FAILED! => {\"changed\": false, \"msg\": \"Always fail\"}\x1b[0m\n\r\nPLAY RECAP *********************************************************************\r\n\x1b[0;31mlocalhost\x1b[0m                  : ok=0    changed=0    unreachable=0    \x1b[0;31mfailed=1   \x1b[0m skipped=0    rescued=0    ignored=0   \r\n\n"))
		})

		It("determines stdout from an incomplete run", func() {
			events := loadFile("./test-events3.jsonl")
			stdout := GetStdout(events, nil)
			Expect(stdout).To(Equal("\r\nPLAY [ping] ********************************************************************\n\r\nTASK [ping] ********************************************************************\n"))
		})

		It("determines stdout from an incomplete run (2)", func() {
			events := loadFile("./test-events4.jsonl")
			stdout := GetStdout(events, nil)
			Expect(stdout).To(Equal(""))
		})
	})
})
