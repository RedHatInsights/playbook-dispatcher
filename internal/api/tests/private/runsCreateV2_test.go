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

func dispatchV2(payload *ApiInternalV2RunsCreateJSONRequestBody) (*RunsCreated, *ApiInternalV2RunsCreateResponse) {
	resp, err := client.ApiInternalV2RunsCreate(test.TestContext(), *payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalV2RunsCreateResponse(resp)
	Expect(err).ToNot(HaveOccurred())
	Expect(res.StatusCode()).To(Equal(http.StatusMultiStatus))

	return res.JSON207, res
}

func minimalV2Payload(recipient uuid.UUID) RunInputV2 {
	return RunInputV2{
		Recipient: public.RunRecipient(recipient.String()),
		Url:       public.Url("http://example.com"),
		OrgId:     public.OrgId("5318290"),
		Name:      public.PlaybookName("ansible playbook"),
		Principal: Principal("test_user"),
	}
}

var _ = Describe("runsCreate V2", func() {
	db := test.WithDatabase()

	It("sends 400 for invalid orgId", func() {
		payload := minimalV2Payload(uuid.New())
		payload.OrgId = "1234"

		runs, _ := dispatchV2(&ApiInternalV2RunsCreateJSONRequestBody{payload})
		Expect((*runs)[0].Code).To(Equal(400))
	})

	It("creates a new ansible playbook run", func() {
		payload := minimalV2Payload(uuid.New())
		payload.OrgId = "12900172"

		runs, _ := dispatchV2(&ApiInternalV2RunsCreateJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(201))
		_, err := uuid.Parse(string(*(*runs)[0].Id))
		Expect(err).ToNot(HaveOccurred())

		var run dbModel.Run
		result := db().Where("id = ?", string(*(*runs)[0].Id)).First(&run)
		Expect(result.Error).ToNot(HaveOccurred())
		Expect(run.Account).To(Equal("12900-test"))
		Expect(run.OrgID).To(Equal(string(payload.OrgId)))
		Expect(run.Recipient.String()).To(Equal(string(payload.Recipient)))
		Expect(run.URL).To(Equal(string(payload.Url)))

		Expect(*run.PlaybookName).To(Equal(string(payload.Name)))
		Expect(run.Status).To(Equal("running"))
		Expect(run.Labels).To(BeEmpty())
		Expect(run.Timeout).To(Equal(3600))
	})

	It("creates a new satellite playbook run", func() {
		recipient := uuid.New()
		url := "http://example.com"
		orgId := "5318290"

		playbookName := public.PlaybookName("sat-playbook")
		playbookRunUrl := public.WebConsoleUrl("http://example.com/webConsoleUrl")
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
		Expect(run.Account).To(Equal("53182-test"))
		Expect(run.OrgID).To(Equal(orgId))
		Expect(run.Recipient).To(Equal(recipient))
		Expect(run.URL).To(Equal(url))
		Expect(run.Status).To(Equal("running"))
		Expect(run.Labels).To(BeEmpty())
		Expect(run.Timeout).To(Equal(3600))

		Expect(*run.PlaybookName).To(Equal(string(playbookName)))
		Expect(run.PlaybookRunUrl).To(Equal(string(playbookRunUrl)))
		Expect(*run.Principal).To(Equal(string(principal)))
		Expect(*run.SatId).To(Equal(satId))
		Expect(*run.SatOrgId).To(Equal(satOrgId))
	})

	It("adds inventory id into hosts column for satellite", func() {
		recipient := uuid.New()
		url := "http://example.com"
		orgId := "5318290"

		playbookName := public.PlaybookName("sat-playbook")
		playbookRunUrl := public.WebConsoleUrl("http://example.com/webConsoleUrl")
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

		var runHost dbModel.RunHost
		result := db().Where("run_id = ?", string(*(*runs)[0].Id)).First(&runHost)
		Expect(result.Error).ToNot(HaveOccurred())

		Expect(runHost.Host).To(Equal(inventoryId))
		Expect(runHost.InventoryID.String()).To(Equal(inventoryId))
	})

	It("creates a new satellite playbook run with a uuidv5 as the sat_id", func() {
		recipient := uuid.MustParse("9200e4a3-c97c-4021-9856-82fa4673e8d2") // gets checked my cloud connector mock
		url := "http://example.com"
		orgId := "5318290"

		playbookName := public.PlaybookName("sat-playbook")
		playbookRunUrl := public.WebConsoleUrl("http://example.com/webConsoleUrl")
		principal := Principal("test_user")

		satOrgId := "123"
		satIdStringV5 := "9274c274-a258-5d00-91fe-dbe0f7849cef"

		inventoryId := uuid.New().String()

		payload := ApiInternalV2RunsCreateJSONRequestBody{
			RunInputV2{
				Recipient:       public.RunRecipient(recipient.String()),
				OrgId:           public.OrgId(orgId),
				Url:             public.Url(url),
				Hosts:           &RunInputHosts{{InventoryId: &inventoryId}},
				Name:            playbookName,
				WebConsoleUrl:   &playbookRunUrl,
				Principal:       principal,
				RecipientConfig: &RecipientConfig{SatId: &satIdStringV5, SatOrgId: &satOrgId},
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
		Expect((*run.SatId).String()).To(Equal(satIdStringV5))
		Expect(*run.SatOrgId).To(Equal(satOrgId))
	})

	It("fails on sat_id mismatch in cloud connector", func() {
		recipient := uuid.MustParse("9200e4a3-c97c-4021-9856-82fa4673e8d2")
		url := "http://example.com"
		orgId := "5318290"

		playbookName := public.PlaybookName("sat-playbook")
		playbookRunUrl := public.WebConsoleUrl("http://example.com/webConsoleUrl")
		principal := Principal("test_user")

		satOrgId := "123"
		satIdStringV5 := "562daa36-b5d8-5511-8bb2-18095e477978" // mismached sat_id recipient combo

		inventoryId := uuid.New().String()

		payload := ApiInternalV2RunsCreateJSONRequestBody{
			RunInputV2{
				Recipient:       public.RunRecipient(recipient.String()),
				OrgId:           public.OrgId(orgId),
				Url:             public.Url(url),
				Hosts:           &RunInputHosts{{InventoryId: &inventoryId}},
				Name:            playbookName,
				WebConsoleUrl:   &playbookRunUrl,
				Principal:       principal,
				RecipientConfig: &RecipientConfig{SatId: &satIdStringV5, SatOrgId: &satOrgId},
			},
		}

		runs, _ := dispatchV2(&payload)

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(500))
	})

	It("sets default for webConsoleUrl", func() {
		payload := minimalV2Payload(uuid.New())

		runs, _ := dispatchV2(&ApiInternalV2RunsCreateJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(201))
		_, err := uuid.Parse(string(*(*runs)[0].Id))
		Expect(err).ToNot(HaveOccurred())

		var run dbModel.Run
		result := db().Where("id = ?", string(*(*runs)[0].Id)).First(&run)
		Expect(result.Error).ToNot(HaveOccurred())

		Expect(run.PlaybookRunUrl).To(Equal(webConsoleUrlDefault))
	})

	It("404s if the recipient is not known", func() {
		payload := minimalV2Payload(uuid.MustParse("b5fbb740-5590-45a4-8240-89192dc49199"))

		runs, _ := dispatchV2(&ApiInternalV2RunsCreateJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(404))
	})

	It("400s if tenant is not known", func() {
		payload := minimalV2Payload(uuid.MustParse("b31955fb-3064-4f56-ae44-a1c488a28587"))
		payload.OrgId = "654321"

		runs, _ := dispatchV2(&ApiInternalV2RunsCreateJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(400))
	})

	It("400s on anemic tenant", func() {
		payload := minimalV2Payload(uuid.MustParse("b31955fb-3064-4f56-ae44-a1c488a28587"))
		payload.OrgId = "654322"

		runs, _ := dispatchV2(&ApiInternalV2RunsCreateJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(400))
	})

	It("500s on cloud connector error", func() {
		payload := minimalV2Payload(uuid.MustParse("b31955fb-3064-4f56-ae44-a1c488a28587"))

		runs, _ := dispatchV2(&ApiInternalV2RunsCreateJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(500))
	})

	It("stores the principal as owning service", func() {
		payload := ApiInternalV2RunsCreateJSONRequestBody{
			minimalV2Payload(uuid.New()),
		}

		ctx := context.WithValue(test.TestContext(), pskKey, "9yh9WuXWDj") //nolint:staticcheck
		resp, err := client.ApiInternalV2RunsCreate(ctx, payload)
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
		payload := ApiInternalV2RunsCreateJSONRequestBody{
			minimalV2Payload(uuid.New()),
		}

		start := time.Now()
		// send 10 requests
		for i := 0; i < 10; i++ {
			_, err := client.ApiInternalV2RunsCreate(test.TestContext(), payload)
			Expect(err).ToNot(HaveOccurred())
		}
		end := time.Since(start)

		Expect(end).To(BeNumerically(">=", time.Second))
	})

	DescribeTable("validation",
		func(payload, expected string) {
			resp, err := client.ApiInternalV2RunsCreateWithBody(test.TestContext(), "application/json", strings.NewReader(payload))
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
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "recipient_config": {"sat_id": "e7ee7fdd-b732-4eea-a070-40025b3dddd9"}}]`,
			`Both sat_id and sat_org need to be defined`,
		),
		Entry(
			"sat_org_id missing",
			`[{"recipient": "3831fec2-1875-432a-bb58-08e71908f0e6", "org_id": "5318290", "principal": "test-user", "url": "http://example.com", "name": "Red Hat Playbook", "recipient_config": {"sat_org_id": "1"}}]`,
			`Both sat_id and sat_org need to be defined`,
		),
	)
})
