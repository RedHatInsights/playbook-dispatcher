package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/controllers/public"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils/test"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func cancelV2(payload *ApiInternalV2RunsCancelJSONRequestBody) (*RunsCanceled, *ApiInternalV2RunsCancelResponse) {
	resp, err := client.ApiInternalV2RunsCancel(test.TestContext(), *payload)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiInternalV2RunsCancelResponse(resp)
	Expect(err).ToNot(HaveOccurred())
	Expect(res.StatusCode()).To(Equal(http.StatusMultiStatus))

	return res.JSON207, res
}

func minimalV2Cancel() CancelInputV2 {
	return CancelInputV2{
		OrgId:     OrgId("5318290"),
		Principal: Principal("test_user"),
		RunId:     public.RunId("eb05b1a7-6613-442f-9ca6-1d7f83188915"),
	}
}

var _ = Describe("runsCancel V2", func() {
	db := test.WithDatabase()

	It("creates cancelation message on successful operation", func() {
		satId := uuid.MustParse("95cbea43-bb85-4153-96c2-eb2474b3e2b3")
		satOrgId := "2"

		var data = test.NewRun(orgId())
		data.Labels = dbModel.Labels{"foo": "bar"}
		data.Timeout = 600
		data.SatId = &satId
		data.SatOrgId = &satOrgId
		Expect(db().Create(&data).Error).ToNot(HaveOccurred())

		payload := minimalV2Cancel()
		payload.RunId = public.RunId(data.ID.String())
		payload.OrgId = "12900172"

		runs, _ := cancelV2(&ApiInternalV2RunsCancelJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(202))
		parsedRunID, err := uuid.Parse(string((*runs)[0].RunId))
		Expect(parsedRunID).To(BeEquivalentTo(data.ID))
		Expect(err).ToNot(HaveOccurred())
	})

	It("404s if playbook run is not known", func() {
		payload := minimalV2Cancel()
		payload.OrgId = "12900172"

		runs, _ := cancelV2(&ApiInternalV2RunsCancelJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(404))
	})

	It("Successfully handles an anemic tenant", func() {
		satId := uuid.New()
		satOrgId := "2"

		var data = test.NewRun(orgId())
		data.Labels = dbModel.Labels{"foo": "bar"}
		data.Timeout = 600
		data.SatId = &satId
		data.SatOrgId = &satOrgId
		Expect(db().Create(&data).Error).ToNot(HaveOccurred())

		payload := minimalV2Cancel()
		payload.RunId = public.RunId(data.ID.String())
		payload.OrgId = "654322"

		runs, _ := cancelV2(&ApiInternalV2RunsCancelJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(202))
		parsedRunID, err := uuid.Parse(string((*runs)[0].RunId))
		Expect(parsedRunID).To(BeEquivalentTo(data.ID))
		Expect(err).ToNot(HaveOccurred())
	})

	It("400s if run is not of type satellite RHC", func() {
		var data = test.NewRun(orgId())
		data.Labels = dbModel.Labels{"foo": "bar"}
		data.Timeout = 600
		Expect(db().Create(&data).Error).ToNot(HaveOccurred())

		payload := minimalV2Cancel()
		payload.RunId = public.RunId(data.ID.String())
		payload.OrgId = "12900172"

		runs, _ := cancelV2(&ApiInternalV2RunsCancelJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(400))
	})

	It("409s on the run being completed", func() {
		satId, _ := uuid.Parse("95cbea43-bb85-4153-96c2-eb2474b3e2b3")
		satOrgId := "2"

		var data = test.NewRun(orgId())
		data.Labels = dbModel.Labels{"foo": "bar"}
		data.Timeout = 600
		data.SatId = &satId
		data.SatOrgId = &satOrgId
		data.Status = "failure"
		Expect(db().Create(&data).Error).ToNot(HaveOccurred())

		payload := minimalV2Cancel()
		payload.RunId = public.RunId(data.ID.String())
		payload.OrgId = "12900172"

		runs, _ := cancelV2(&ApiInternalV2RunsCancelJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(409))
	})

	It("500s on cloud connector error", func() {
		var data = test.NewRun(orgId())
		var newRecipient, _ = uuid.Parse("b31955fb-3064-4f56-ae44-a1c488a28587")
		satId, _ := uuid.Parse("95cbea43-bb85-4153-96c2-eb2474b3e2b3")
		satOrgId := "2"

		data.Recipient = newRecipient
		data.Labels = dbModel.Labels{"foo": "bar"}
		data.Timeout = 600
		data.SatId = &satId
		data.SatOrgId = &satOrgId
		Expect(db().Create(&data).Error).ToNot(HaveOccurred())

		payload := minimalV2Cancel()
		payload.RunId = public.RunId(data.ID.String())
		payload.OrgId = "12900172"

		runs, _ := cancelV2(&ApiInternalV2RunsCancelJSONRequestBody{payload})

		Expect(*runs).To(HaveLen(1))
		Expect((*runs)[0].Code).To(Equal(500))
	})
})
