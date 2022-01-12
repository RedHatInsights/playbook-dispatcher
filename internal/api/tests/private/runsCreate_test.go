package private

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	p "playbook-dispatcher/internal/api/controllers/private"
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

func dispatchV2(payload *ApiInternalV2RunsCreateJSONRequestBody) (*RunsCreated, *ApiInternalV2RunsCreateResponse) {
	resp, err := client.ApiInternalV2RunsCreate(test.TestContext(), *payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalV2RunsCreateResponse(resp)
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

			ctx := context.WithValue(test.TestContext(), pskKey, "9yh9WuXWDj")
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

			ctx := context.WithValue(test.TestContext(), pskKey, "9yh9WuXWDj")
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

	Describe("create run v2", func() {
		db := test.WithDatabase()

		It("sends 404 for invalid orgId", func() {
			recipient := uuid.New()
			url := "http://example.com"
			orgId := "1234"

			playbookName := PlaybookName("ansible-playbook")
			principal := Principal("test_user")

			payload := ApiInternalV2RunsCreateJSONRequestBody{
				RunInputV2{
					Recipient: public.RunRecipient(recipient.String()),
					Url:       public.Url(url),
					OrgId:     public.OrgId(orgId),
					Name:      playbookName,
					Principal: principal,
				},
			}

			runs, _ := dispatchV2(&payload)
			Expect((*runs)[0].Code).To(Equal(400))
		})

		It("creates a new ansible playbook run", func() {
			recipient := uuid.New()
			url := "http://example.com"
			orgId := "12900172"

			playbookName := PlaybookName("ansible-playbook")
			principal := Principal("test_user")

			payload := ApiInternalV2RunsCreateJSONRequestBody{
				RunInputV2{
					Recipient: public.RunRecipient(recipient.String()),
					Url:       public.Url(url),
					OrgId:     public.OrgId(orgId),
					Name:      playbookName,
					Principal: principal,
					Hosts:     &RunInputHosts{{AnsibleHost: &ansibleHost}},
				},
			}

			runs, _ := dispatchV2(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(201))
			_, err := uuid.Parse(string(*(*runs)[0].Id))
			Expect(err).ToNot(HaveOccurred())

			var run dbModel.Run
			result := db().Where("id = ?", string(*(*runs)[0].Id)).First(&run)
			Expect(result.Error).ToNot(HaveOccurred())
			Expect(run.Account).To(Equal("6377882"))
			Expect(run.OrgID).To(Equal(orgId))
			Expect(run.Recipient).To(Equal(recipient))
			Expect(run.URL).To(Equal(url))

			Expect(run.PlaybookName).To(Equal(string(playbookName)))
			Expect(run.Status).To(Equal("running"))
			Expect(run.Labels).To(BeEmpty())
			Expect(run.Timeout).To(Equal(3600))
		})

		It("creates a new satellite playbook run", func() {
			recipient := uuid.New()
			url := "http://example.com"
			orgId := "5318290"

			playbookName := PlaybookName("sat-playbook")
			playbookRunUrl := WebConsoleUrl("http://example.com/webConsoleUrl")
			principal := Principal("test_user")

			satId := uuid.New()
			satOrgId := "123"
			inventoryId := uuid.New().String()
			satIdString := satId.String()

			payload := ApiInternalV2RunsCreateJSONRequestBody{
				RunInputV2{
					Recipient:       public.RunRecipient(recipient.String()),
					OrgId:           public.OrgId(orgId),
					Url:             public.Url(url),
					Hosts:           &RunInputHosts{{InventoryId: &inventoryId}},
					Name:            playbookName,
					WebConsoleUrl:   &playbookRunUrl,
					Principal:       principal,
					RecipientConfig: &RecipientConfig{SatId: &satIdString, SatOrgId: &satOrgId},
				},
			}

			runs, _ := dispatchV2(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(201))
			_, err := uuid.Parse(string(*(*runs)[0].Id))
			Expect(err).ToNot(HaveOccurred())

			var run dbModel.Run
			result := db().Where("id = ?", string(*(*runs)[0].Id)).First(&run)
			Expect(result.Error).ToNot(HaveOccurred())
			Expect(run.Account).To(Equal("901578"))
			Expect(run.OrgID).To(Equal(orgId))
			Expect(run.Recipient).To(Equal(recipient))
			Expect(run.URL).To(Equal(url))
			Expect(run.Status).To(Equal("running"))
			Expect(run.Labels).To(BeEmpty())
			Expect(run.Timeout).To(Equal(3600))

			Expect(run.PlaybookName).To(Equal(string(playbookName)))
			Expect(run.PlaybookRunUrl).To(Equal(string(playbookRunUrl)))
			Expect(run.Principal).To(Equal(string(principal)))
			Expect(run.SatId).To(Equal(satId))
			Expect(run.SatOrgId).To(Equal(satOrgId))
		})

		It("sets default for webConsoleUrl", func() {
			recipient := uuid.New()
			url := "http://example.com"
			orgId := "5318290"

			playbookName := PlaybookName("playbook")
			principal := Principal("test_user")

			payload := ApiInternalV2RunsCreateJSONRequestBody{
				RunInputV2{
					Recipient: public.RunRecipient(recipient.String()),
					Url:       public.Url(url),
					OrgId:     public.OrgId(orgId),
					Name:      playbookName,
					Principal: principal,
				},
			}

			runs, _ := dispatchV2(&payload)

			Expect(*runs).To(HaveLen(1))
			Expect((*runs)[0].Code).To(Equal(201))
			_, err := uuid.Parse(string(*(*runs)[0].Id))
			Expect(err).ToNot(HaveOccurred())

			var run dbModel.Run
			result := db().Where("id = ?", string(*(*runs)[0].Id)).First(&run)
			Expect(result.Error).ToNot(HaveOccurred())

			Expect(run.PlaybookRunUrl).To(Equal(webConsoleUrlDefault))
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

	DescribeTable("checkReqField",
		func(runInputJson string, output []bool) {
			runInput := p.RunInputV2{}
			byteString := []byte(runInputJson)

			err := json.Unmarshal(byteString, &runInput)
			Expect(err).To(BeNil())

			ansible, satellite := p.CheckV2ReqFields(runInput)

			Expect(ansible).To(Equal(output[0]))
			Expect(satellite).To(Equal(output[1]))
		},

		Entry("none valid", `{"org_id": "123", "hosts": [{"inventory_id": "16372e6f-1c18-4cdb-b780-50ab4b88e74b"}]}`, []bool{false, false}),
		Entry(
			"valid ansible",
			`{"org_id": "123", "recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c", "url": "http://example.com", "principal": "test-user","hosts": [{"ansible_host": "localhost","inventory_id": "16372e6f-1c18-4cdb-b780-50ab4b88e74b"}]}`,
			[]bool{true, false},
		),
		Entry("valid satellite",
			`{"org_id": "123","recipient": "dd018b96-da04-4651-84d1-187fa5c23f6c","url": "http://example.com","playbook_name":"test-playbook","playbook_run_url": "http://example.com","recipient_config": {"sat_id":"16372e6f-1c18-4cdb-b780-50ab4b88e74b","sat_org_id":"456"},"principal": "test-user","hosts": [{"inventory_id": "16372e6f-1c18-4cdb-b780-50ab4b88e74b"}, {"inventory_id": "baf2bb2f-06a3-42cc-ae7b-68ccc8e2a344"}]}`,
			[]bool{false, true},
		),
	)

})
