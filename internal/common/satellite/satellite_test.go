package satellite

import (
	"encoding/json"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func loadFile(path string) (events []messageModel.PlaybookSatRunResponseMessageYamlEventsElem) {
	file, err := os.ReadFile(path)
	Expect(err).ToNot(HaveOccurred())

	lines := strings.Split(string(file), "\n")

	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		event := messageModel.PlaybookSatRunResponseMessageYamlEventsElem{}
		err = json.Unmarshal([]byte(line), &event)
		Expect(err).ToNot(HaveOccurred())

		events = append(events, event)
	}

	return events
}

var _ = Describe("Satellite", func() {
	Describe("host", func() {
		It("determines satellite host from a successful run", func() {
			events := loadFile("./sat-test-events1.jsonl")
			hosts := GetSatHosts(events)
			Expect(hosts).To(HaveLen(1))
			Expect(hosts[0]).To(Equal("2798f2ab-35b2-4d4e-af0c-0478dcb4a324"))
		})

		It("determines satellite host from a failed run", func() {
			events := loadFile("./sat-test-events2.jsonl")
			hosts := GetSatHosts(events)
			Expect(hosts).To(HaveLen(1))
			Expect(hosts[0]).To(Equal("2798f2ab-35b2-4d4e-af0c-0478dcb4a324"))
		})

		It("determines satellite host from a cancelled run", func() {
			events := loadFile("./sat-test-events3.jsonl")
			hosts := GetSatHosts(events)
			Expect(hosts).To(HaveLen(1))
			Expect(hosts[0]).To(Equal("2798f2ab-35b2-4d4e-af0c-0478dcb4a324"))
		})

		It("determines satellite hosts from a multi-host run", func() {
			events := loadFile("./sat-test-events4.jsonl")
			hosts := GetSatHosts(events)
			Expect(hosts).To(HaveLen(2))
			Expect(hosts[0]).To(Equal("2798f2ab-35b2-4d4e-af0c-0478dcb4a324"))
			Expect(hosts[1]).To(Equal("e8f05d27-5af9-4547-a30c-5be04e099ffe"))
		})
	})

	Describe("satHostInfo", func() {
		It("determines satHostinfo from a run", func() {
			events := loadFile("./sat-test-events2.jsonl")
			host := "2798f2ab-35b2-4d4e-af0c-0478dcb4a324"
			satHostInfo := GetSatHostInfo(events, &host)
			Expect(*satHostInfo.Sequence).To(Equal(1))
			Expect(satHostInfo.Console).To(Equal("localhost | FAILED => {\n    \"changed\": false,\n    \"ping\": \"runtime_error\"\n}"))
		})

		It("determines correct satHostinfo from a multi-host run", func() {
			events := loadFile("./sat-test-events4.jsonl")
			host := "e8f05d27-5af9-4547-a30c-5be04e099ffe"
			satHostInfo := GetSatHostInfo(events, &host)
			Expect(*satHostInfo.Sequence).To(Equal(4))
			Expect(satHostInfo.Console).To(Equal("host2 | SUCCESS => {\n    \"changed\": false,\n    \"ping\": \"pong\"\n}"))
		})
	})
})
