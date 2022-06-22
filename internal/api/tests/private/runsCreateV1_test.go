package private

import (
	"context"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/api/controllers/public"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils/test"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var (
	ansibleHost = "localhost"
)

func dispatch(payload *ApiInternalRunsCreateJSONRequestBody) (*RunsCreated, *ApiInternalRunsCreateResponse) {
	resp, err := client.ApiInternalRunsCreate(test.TestContext(), *payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalRunsCreateResponse(resp)
	Expect(err).ToNot(HaveOccurred())
	Expect(res.StatusCode()).To(Equal(http.StatusMultiStatus))

	return res.JSON207, res
}

var _ = Describe("runsCreate V1", func() {
	Describe("create run happy path", func() {
		db := test.WithDatabase()

		It("creates a new playbook run", func() {
			recipient := uuid.New()
			url := "http://example.com"

			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: public.RunRecipient(recipient.String()),
					Account:   public.Account(accountNumber()),
					Url:       public.Url(url),
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			runs, _ := dispatch(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(201))
			_, err := uuid.Parse(string(*(*runs)[0].Id))
			Expect(err).ToNot(HaveOccurred())

			var run dbModel.Run
			result := db().Where("id = ?", string(*(*runs)[0].Id)).First(&run)
			Expect(result.Error).ToNot(HaveOccurred())
			Expect(run.Account).To(Equal(accountNumber()))
			Expect(run.Recipient).To(Equal(recipient))
			Expect(run.URL).To(Equal(url))
			Expect(run.Status).To(Equal("running"))
			Expect(run.Labels).To(BeEmpty())
			Expect(run.Timeout).To(Equal(3600))
		})

		It("404s if the recipient is not known", func() {
			recipient := uuid.MustParse("b5fbb740-5590-45a4-8240-89192dc49199")
			url := "http://example.com"

			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: public.RunRecipient(recipient.String()),
					Account:   public.Account(accountNumber()),
					Url:       public.Url(url),
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			runs, _ := dispatch(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(404))
		})

		It("Populates missing OrgID for incoming runs", func() {
			recipient := uuid.New()
			url := "http://example.com"

			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: public.RunRecipient(recipient.String()),
					Account:   public.Account("10000"),
					Url:       public.Url(url),
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			runs, _ := dispatch(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(201))

			var run dbModel.Run
			result := db().Where("id = ?", string(*(*runs)[0].Id)).First(&run)
			Expect(result.Error).ToNot(HaveOccurred())
			Expect(run.OrgID).To(Equal("10000-test"))
			Expect(run.Account).To(Equal("10000"))
			Expect(run.Recipient).To(Equal(recipient))
			Expect(run.URL).To(Equal(url))
			Expect(run.Status).To(Equal("running"))
			Expect(run.Labels).To(BeEmpty())
			Expect(run.Timeout).To(Equal(3600))

		})

		It("404s if the OrgID is not found", func() {
			recipient := uuid.New()
			url := "http://example.com"

			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: public.RunRecipient(recipient.String()),
					Account:   public.Account("1234"),
					Url:       public.Url(url),
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			runs, _ := dispatch(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(404))
		})

		It("500s on cloud connector error", func() {
			recipient := uuid.MustParse("b31955fb-3064-4f56-ae44-a1c488a28587")
			url := "http://example.com"

			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: public.RunRecipient(recipient.String()),
					Account:   public.Account(accountNumber()),
					Url:       public.Url(url),
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			runs, _ := dispatch(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(500))
		})

		It("stores the principal as owning service", func() {
			recipient := uuid.New()
			url := "http://example.com"
			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: public.RunRecipient(recipient.String()),
					Account:   public.Account(accountNumber()),
					Url:       public.Url(url),
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			ctx := context.WithValue(test.TestContext(), pskKey, "9yh9WuXWDj") //nolint:staticcheck
			resp, err := client.ApiInternalRunsCreate(ctx, payload)
			Expect(err).ToNot(HaveOccurred())
			res, err := ParseApiInternalRunsCreateResponse(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode()).To(Equal(http.StatusMultiStatus))

			runs := *res.JSON207
			Expect(runs).To(HaveLen(1))
			Expect(runs[0].Code).To(Equal(201))

			var run dbModel.Run
			result := db().Where("id = ?", string(*runs[0].Id)).First(&run)
			Expect(result.Error).ToNot(HaveOccurred())
			Expect(run.Service).To(Equal("test02"))
		})

		It("enforces rate limit", func() {
			recipient := uuid.New()
			url := "http://example.com"

			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: public.RunRecipient(recipient.String()),
					Account:   public.Account(accountNumber()),
					Url:       public.Url(url),
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			ctx := context.WithValue(test.TestContext(), pskKey, "9yh9WuXWDj") //nolint:staticcheck
			start := time.Now()
			// send 10 requests
			for i := 0; i < 10; i++ {
				_, err := client.ApiInternalRunsCreate(ctx, payload)
				Expect(err).ToNot(HaveOccurred())
			}
			end := time.Since(start)

			Expect(end).To(BeNumerically(">=", time.Second))
		})
	})

	DescribeTable("validation",
		func(payload, expected string) {
			resp, err := client.ApiInternalRunsCreateWithBody(test.TestContext(), "application/json", strings.NewReader(payload))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(ContainSubstring(expected))
		},

		Entry("empty list", `[]`, "Minimum number of items is 1"),
		Entry(
			"missing required property (account)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "url": "http://example.com"}]`,
			"Property 'account' is missing",
		),
		Entry(
			"invalid property (account)",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "url": "http://example.com", "account": "2718281828459045235360287471352"}]`,
			"Maximum string length is 10",
		),
		Entry(
			"timeout minimum",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "url": "http://example.com", "account": "540155", "timeout": -1}]`,
			"Number must be at least 0",
		),
		Entry(
			"timeout maximum",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "url": "http://example.com", "account": "540155", "timeout": 1000000}]`,
			"Number must be most 604800",
		),
	)
})
