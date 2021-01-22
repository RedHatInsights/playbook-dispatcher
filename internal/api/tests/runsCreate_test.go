package tests

import (
	"context"
	"io/ioutil"
	"net/http"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils/test"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func dispatch(payload *ApiInternalRunsCreateJSONRequestBody) (*RunsCreated, *ApiInternalRunsCreateResponse) {
	resp, err := client.ApiInternalRunsCreate(context.Background(), *payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalRunsCreateResponse(resp)
	Expect(err).ToNot(HaveOccurred())
	Expect(res.StatusCode()).To(Equal(http.StatusMultiStatus))

	return res.JSON207, res
}

var _ = Describe("runsCreate", func() {
	Describe("create run happy path", func() {
		db := test.WithDatabase()

		It("creates a new playbook run", func() {
			recipient := uuid.New()
			url := "http://example.com"
			payload := ApiInternalRunsCreateJSONRequestBody{
				RunInput{
					Recipient: RunRecipient(recipient.String()),
					Account:   Account(accountNumber()),
					Url:       Url(url),
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
			Expect(run.PlaybookURL).To(Equal(url))
			Expect(run.Status).To(Equal("running"))
			Expect(run.Labels).To(BeEmpty())
			Expect(run.Timeout).To(Equal(3600))
		})
	})

	DescribeTable("validation",
		func(payload, expected string) {
			resp, err := client.ApiInternalRunsCreateWithBody(context.Background(), "application/json", strings.NewReader(payload))
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
	)

})
