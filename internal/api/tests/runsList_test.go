package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils/test"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func listRuns(keysAndValues ...interface{}) (*Runs, *ApiRunsListResponse) {
	raw := listRunsRaw(keysAndValues...)
	res, err := ParseApiRunsListResponse(raw)
	Expect(err).ToNot(HaveOccurred())

	return res.JSON200, res
}

func listRunsRaw(keysAndValues ...interface{}) *http.Response {
	if len(keysAndValues)%2 != 0 {
		panic("Odd number of parameters")
	}

	requestUrl := "http://localhost:9002/api/playbook-dispatcher/v1/runs"

	params := make([]string, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues)/2; i++ {
		params[i] = fmt.Sprintf("%s=%s",
			url.QueryEscape(fmt.Sprintf("%s", keysAndValues[i*2])),
			url.QueryEscape(fmt.Sprintf("%v", keysAndValues[(i*2)+1])),
		)
	}

	query := strings.Join(params, "&")
	requestUrl += fmt.Sprintf("?%s", query)

	req, err := http.NewRequest("GET", requestUrl, nil)
	Expect(err).ToNot(HaveOccurred())
	req.Header.Set("x-rh-identity", test.IdentityHeaderMinimal(accountNumber()))
	resp, err := test.Client.Do(req)
	Expect(err).ToNot(HaveOccurred())
	return resp
}

var _ = Describe("runsList", func() {
	db := test.WithDatabase()

	Describe("list runs", func() {
		It("by default returns a list of existing runs", func() {
			var data = test.NewRunWithStatus(accountNumber(), "success")
			data.Labels = dbModel.Labels{"foo": "bar"}
			data.Timeout = 600
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			runs, res := listRuns()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Data).To(HaveLen(1))
			run := runs.Data[0]
			Expect(string(*run.Id)).To(Equal(data.ID.String()))
			Expect(run.Labels.AdditionalProperties["foo"]).To(Equal(data.Labels["foo"]))
			Expect(string(*run.Recipient)).To(Equal(data.Recipient.String()))
			Expect(string(*run.Status)).To(Equal(data.Status))
			Expect(int(*run.Timeout)).To(Equal(data.Timeout))
			Expect(string(*run.Url)).To(Equal(data.URL))
		})

		It("properly infers run status", func() {
			var data = []dbModel.Run{
				*test.NewRunWithStatus(accountNumber(), "running"),
				*test.NewRunWithStatus(accountNumber(), "success"),
				*test.NewRunWithStatus(accountNumber(), "failure"),
				*test.NewRunWithStatus(accountNumber(), "running"),
			}

			data[1].CreatedAt = time.Date(2020, time.January, 2, 12, 45, 3, 0, time.UTC)
			data[2].CreatedAt = time.Date(2020, time.January, 2, 11, 45, 3, 0, time.UTC)
			data[3].CreatedAt = time.Date(2020, time.January, 2, 10, 45, 3, 0, time.UTC)

			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			runs, res := listRuns()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(string(*runs.Data[0].Status)).To(Equal("running"))
			Expect(string(*runs.Data[1].Status)).To(Equal("success"))
			Expect(string(*runs.Data[2].Status)).To(Equal("failure"))
			Expect(string(*runs.Data[3].Status)).To(Equal("timeout"))
		})
	})

	Describe("sorting", func() {
		BeforeEach(func() {
			var runs = []dbModel.Run{
				*test.NewRunWithStatus(accountNumber(), "success"),
				*test.NewRunWithStatus(accountNumber(), "failure"),
			}

			runs[0].CreatedAt = time.Date(2020, time.January, 21, 8, 45, 3, 0, time.UTC)
			runs[1].CreatedAt = time.Date(2020, time.January, 21, 9, 45, 3, 0, time.UTC)

			Expect(db().Create(&runs).Error).ToNot(HaveOccurred())
		})

		DescribeTable("sorting happy path",
			func(sortBy RunsSortBy, expected ...RunStatus) {
				runs, res := listRuns("sort_by", sortBy)
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(2))

				for i, status := range expected {
					Expect(*runs.Data[i].Status).To(Equal(status))
				}
			},

			Entry("by default orders by created_at desc", nil, RunStatus_failure, RunStatus_success),
			Entry("sorts by created_at", RunsSortBy_created_at, RunStatus_failure, RunStatus_success),
			Entry("sorts by created_at:desc", RunsSortBy_created_at_desc, RunStatus_failure, RunStatus_success),
			Entry("sorts by created_at:asc", RunsSortBy_created_at_asc, RunStatus_success, RunStatus_failure),
		)

		It("400s on invalid value", func() {
			_, res := listRuns("sort_by", "salad:asc")
			Expect(res.StatusCode()).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("pagination", func() {
		BeforeEach(func() {
			var runs = []dbModel.Run{
				*test.NewRun(accountNumber()),
				*test.NewRun(accountNumber()),
				*test.NewRun(accountNumber()),
				*test.NewRun(accountNumber()),
				*test.NewRun(accountNumber()),
			}

			Expect(db().Create(&runs).Error).ToNot(HaveOccurred())
		})

		DescribeTable("pagination happy path",
			func(limit, offset, expected int) {
				runs, res := listRuns(
					"limit", limit,
					"offset", offset,
				)
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
				_, res := listRuns(
					"limit", limit,
					"offset", offset,
				)
				Expect(res.StatusCode()).To(Equal(http.StatusBadRequest))
			},

			Entry("limit=0, offset=0", 0, 0),
			Entry("limit=-1, offset=0", -1, 0),
			Entry("limit=1, offset=-1", -1, -1),
		)

		It("defaults to 50/0", func() {
			runs, res := listRuns()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Meta.Count).To(Equal(5))
		})
	})

	Describe("filtering", func() {
		Describe("status", func() {
			var data []dbModel.Run

			BeforeEach(func() {
				data = []dbModel.Run{
					*test.NewRunWithStatus(accountNumber(), "success"),
					*test.NewRunWithStatus(accountNumber(), "failure"),
					*test.NewRunWithStatus(accountNumber(), "running"),
					*test.NewRunWithStatus(accountNumber(), "running"),
				}

				data[3].CreatedAt = time.Now().Add(-6 * time.Hour)

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())
			})

			DescribeTable("filtering by various status values",
				func(status string, index int) {
					runs, res := listRuns("filter[status]", status)
					Expect(res.StatusCode()).To(Equal(http.StatusOK))
					Expect(runs.Meta.Count).To(Equal(1))
					Expect(string(*runs.Data[0].Id)).To(Equal(data[index].ID.String()))
				},

				Entry("success", "success", 0),
				Entry("failure", "failure", 1),
				Entry("running", "running", 2),
				Entry("timeout", "timeout", 3),
			)
		})

		Describe("recipient", func() {
			var data []dbModel.Run

			BeforeEach(func() {
				data = []dbModel.Run{
					*test.NewRun(accountNumber()),
					*test.NewRun(accountNumber()),
					*test.NewRun(accountNumber()),
					*test.NewRun(accountNumber()),
				}

				data[3].Recipient = uuid.MustParse("64aeb237-d46d-494e-98e3-b48fc5c78bf1")
				data[3].Account = "9999999999"

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())
			})

			It("finds a run based on recipient id", func() {
				runs, res := listRuns("filter[recipient]", data[1].Recipient)
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(1))
				Expect(string(*runs.Data[0].Recipient)).To(Equal(data[1].Recipient.String()))
			})

			It("returns empty result on non-match", func() {
				runs, res := listRuns("filter[recipient]", "b76ceabc-d404-4a43-a09c-7650e661e807")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(0))
			})

			It("respects account isolation", func() {
				runs, res := listRuns("filter[recipient]", "64aeb237-d46d-494e-98e3-b48fc5c78bf1")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(0))
			})
		})

		Describe("labels", func() {
			var data []dbModel.Run

			BeforeEach(func() {
				data = []dbModel.Run{
					*test.NewRun(accountNumber()),
					*test.NewRun(accountNumber()),
					*test.NewRun(accountNumber()),
				}

				data[1].Labels = map[string]string{
					"service": "remediations",
					"foo":     "bar",
				}
				data[2].Labels = map[string]string{
					"service": "remediations",
				}

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())
			})

			It("finds all runs matching given label", func() {
				runs, res := listRuns("filter[labels][service]", "remediations")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(2))
				expectedIds := []string{data[1].ID.String(), data[2].ID.String()}
				Expect(expectedIds).To(ContainElement(string(*runs.Data[0].Id)))
				Expect(expectedIds).To(ContainElement(string(*runs.Data[1].Id)))
			})

			It("finds all runs matching a combination of two labels", func() {
				runs, res := listRuns("filter[labels][service]", "remediations", "filter[labels][foo]", "bar")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(1))
				Expect(string(*runs.Data[0].Id)).To(Equal(data[1].ID.String()))
			})

			It("does not find anything if labels do not match", func() {
				runs, res := listRuns("filter[labels][abc]", "def")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(0))
			})
		})
	})

	Describe("sparse fieldsets", func() {
		DescribeTable("happy path",
			func(fields ...string) {
				Expect(db().Create(test.NewRun(accountNumber())).Error).ToNot(HaveOccurred())

				res := listRunsRaw("fields[data]", strings.Join(fields, ","))
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				bodyBytes, err := ioutil.ReadAll(res.Body)
				Expect(err).ToNot(HaveOccurred())
				defer res.Body.Close()

				representation := make(map[string]interface{})
				json.Unmarshal(bodyBytes, &representation)

				runs := representation["data"].([]interface{})
				Expect(runs[0]).To(HaveLen(len(fields)))
				for _, field := range fields {
					Expect(runs[0]).To(HaveKey(field))
				}
			},

			Entry("single field", "id"),
			Entry("defaults defined explicitly", "id", "recipient", "url", "labels", "timeout", "status"),
			Entry("all fields", "id", "recipient", "url", "labels", "timeout", "status", "created_at", "updated_at"),
		)

		It("400s on invalid value", func() {
			raw := listRunsRaw("fields[data]", "id,salad")
			Expect(raw.StatusCode).To(Equal(http.StatusBadRequest))
			res, err := ParseApiRunsListResponse(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.JSON400.Message).To(Equal("unknown field: salad"))
		})
	})

})
