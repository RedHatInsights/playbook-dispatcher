package tests

import (
	"net/http"
	"playbook-dispatcher/internal/common/models"
	"time"

	"github.com/google/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func listRuns(params *ApiRunsListParams) (*Runs, *ApiRunsListResponse) {
	resp, err := client.ApiRunsList(ContextWithIdentity(accountNumber()), params)
	Expect(err).ToNot(HaveOccurred())
	res, err := ParseApiRunsListResponse(resp)
	Expect(err).ToNot(HaveOccurred())

	return res.JSON200, res
}

func newRunWithStatus(status string) *models.Run {
	return &models.Run{
		ID:            uuid.New(),
		Account:       accountNumber(),
		Recipient:     uuid.New(),
		CorrelationID: uuid.New(),
		PlaybookURL:   "http://example.com",
		Status:        status,
	}
}

func newRun() *models.Run {
	return newRunWithStatus("running")
}

var _ = Describe("runsList", func() {
	db := WithDatabase()

	Describe("list runs", func() {
		It("by default returns a list of existing runs", func() {
			var data = newRunWithStatus("success")
			data.Labels = models.Labels{"foo": "bar"}
			data.Timeout = 600
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			runs, res := listRuns(&ApiRunsListParams{})
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Data).To(HaveLen(1))
			run := runs.Data[0]
			Expect(string(run.Id)).To(Equal(data.ID.String()))
			Expect(string(run.Account)).To(Equal(data.Account))
			Expect(run.Labels.AdditionalProperties["foo"]).To(Equal(data.Labels["foo"]))
			Expect(string(run.Recipient)).To(Equal(data.Recipient.String()))
			Expect(string(run.Status)).To(Equal(data.Status))
			Expect(int(run.Timeout)).To(Equal(data.Timeout))
			Expect(string(run.Url)).To(Equal(data.PlaybookURL))
		})
	})

	Describe("sorting", func() {
		BeforeEach(func() {
			var runs = []models.Run{
				*newRunWithStatus("success"),
				*newRunWithStatus("running"),
			}

			runs[0].CreatedAt = time.Date(2020, time.January, 21, 8, 45, 3, 0, time.UTC)
			runs[1].CreatedAt = time.Date(2020, time.January, 21, 9, 45, 3, 0, time.UTC)

			Expect(db().Create(&runs).Error).ToNot(HaveOccurred())
		})

		DescribeTable("sorting happy path",
			func(sortBy RunsSortBy, expected ...RunStatus) {
				runs, res := listRuns(&ApiRunsListParams{SortBy: &sortBy})
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(2))

				for i, status := range expected {
					Expect(runs.Data[i].Status).To(Equal(status))
				}
			},

			Entry("by default orders by created_at desc", nil, RunStatus_running, RunStatus_success),
			Entry("sorts by created_at", RunsSortBy_created_at, RunStatus_running, RunStatus_success),
			Entry("sorts by created_at:desc", RunsSortBy_created_at_desc, RunStatus_running, RunStatus_success),
			Entry("sorts by created_at:asc", RunsSortBy_created_at_asc, RunStatus_success, RunStatus_running),
		)

		It("400s on invalid value", func() {
			sortBy := RunsSortBy("salad:asc")
			_, res := listRuns(&ApiRunsListParams{SortBy: &sortBy})
			Expect(res.StatusCode()).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("pagination", func() {
		BeforeEach(func() {
			var runs = []models.Run{
				*newRun(),
				*newRun(),
				*newRun(),
				*newRun(),
				*newRun(),
			}

			Expect(db().Create(&runs).Error).ToNot(HaveOccurred())
		})

		DescribeTable("pagination happy path",
			func(limit, offset, expected int) {
				l := Limit(limit)
				o := Offset(offset)
				runs, res := listRuns(&ApiRunsListParams{Limit: &l, Offset: &o})
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(expected))
				Expect(runs.Data).To(HaveLen(expected))
			},

			Entry("limit=1, offset=0", 1, 0, 1),
			Entry("limit=3, offset=0", 3, 0, 3),
			Entry("limit=10, offset=0", 10, 0, 5),
			Entry("limit=1, offset=3", 1, 3, 1),
			Entry("limit=5, offset=3", 5, 3, 2),
			Entry("limit=5, offset=10", 5, 10, 0),
		)

		DescribeTable("pagination invalid values",
			func(limit, offset int) {
				l := Limit(limit)
				o := Offset(offset)
				_, res := listRuns(&ApiRunsListParams{Limit: &l, Offset: &o})
				Expect(res.StatusCode()).To(Equal(http.StatusBadRequest))
			},

			Entry("limit=0, offset=0", 0, 0),
			Entry("limit=-1, offset=0", -1, 0),
			Entry("limit=1, offset=-1", -1, -1),
		)

		It("defaults to 50/0", func() {
			runs, res := listRuns(&ApiRunsListParams{})
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Meta.Count).To(Equal(5))
		})
	})
})
