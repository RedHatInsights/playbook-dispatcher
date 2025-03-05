package public

import (
	"net/http"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils/test"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func listRunsRaw(keysAndValues ...interface{}) *http.Response {
	return doGet("http://localhost:9002/api/playbook-dispatcher/v1/runs", keysAndValues...)
}

func listRuns(keysAndValues ...interface{}) (*Runs, *ApiRunsListResponse) {
	raw := listRunsRaw(keysAndValues...)
	res, err := ParseApiRunsListResponse(raw)
	Expect(err).ToNot(HaveOccurred())

	return res.JSON200, res
}

var _ = Describe("runsList", func() {
	db := test.WithDatabase()

	Describe("list runs", func() {
		It("by default returns a list of existing runs", func() {
			data := test.NewRunWithStatus(orgId(), "success")
			data.Labels = dbModel.Labels{"foo": "bar"}
			data.Timeout = 600
			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			runs, res := listRuns()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Data).To(HaveLen(1))
			run := runs.Data[0]
			Expect(*run.Id).To(BeEquivalentTo(data.ID.String()))
			Expect((*run.Labels)["foo"]).To(Equal(data.Labels["foo"]))
			Expect(*run.Recipient).To(BeEquivalentTo(data.Recipient.String()))
			Expect(*run.Status).To(BeEquivalentTo(data.Status))
			Expect(*run.Timeout).To(BeEquivalentTo(data.Timeout))
			Expect(*run.Url).To(BeEquivalentTo(data.URL))
		})

		It("properly infers run status", func() {
			data := []dbModel.Run{
				test.NewRunWithStatus(orgId(), "running"),
				test.NewRunWithStatus(orgId(), "success"),
				test.NewRunWithStatus(orgId(), "failure"),
				test.NewRunWithStatus(orgId(), "running"),
			}

			data[1].CreatedAt = time.Date(2020, time.January, 2, 12, 45, 3, 0, time.UTC)
			data[2].CreatedAt = time.Date(2020, time.January, 2, 11, 45, 3, 0, time.UTC)
			data[3].CreatedAt = time.Date(2020, time.January, 2, 10, 45, 3, 0, time.UTC)

			Expect(db().Create(&data).Error).ToNot(HaveOccurred())

			runs, res := listRuns()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(*runs.Data[0].Status).To(BeEquivalentTo("running"))
			Expect(*runs.Data[1].Status).To(BeEquivalentTo("success"))
			Expect(*runs.Data[2].Status).To(BeEquivalentTo("failure"))
			Expect(*runs.Data[3].Status).To(BeEquivalentTo("timeout"))
		})
	})

	Describe("sorting", func() {
		BeforeEach(func() {
			runs := []dbModel.Run{
				test.NewRunWithStatus(orgId(), "success"),
				test.NewRunWithStatus(orgId(), "failure"),
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
			runs := []dbModel.Run{
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
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
					test.NewRunWithStatus(orgId(), "success"),
					test.NewRunWithStatus(orgId(), "failure"),
					test.NewRunWithStatus(orgId(), "running"),
					test.NewRunWithStatus(orgId(), "running"),
				}

				data[3].CreatedAt = time.Now().Add(-6 * time.Hour)

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())
			})

			DescribeTable("filtering by various status values",
				func(status string, index int) {
					runs, res := listRuns("filter[status]", status)
					Expect(res.StatusCode()).To(Equal(http.StatusOK))
					Expect(runs.Meta.Count).To(Equal(1))
					Expect(*runs.Data[0].Id).To(BeEquivalentTo(data[index].ID.String()))
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
					test.NewRun(orgId()),
					test.NewRun(orgId()),
					test.NewRun(orgId()),
					test.NewRun(orgId()),
				}

				data[3].Recipient = uuid.MustParse("64aeb237-d46d-494e-98e3-b48fc5c78bf1")
				data[3].OrgID = "9999999999-test"

				Expect(db().Create(&data).Error).ToNot(HaveOccurred())
			})

			It("finds a run based on recipient id", func() {
				runs, res := listRuns("filter[recipient]", data[1].Recipient)
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(1))
				Expect(*runs.Data[0].Recipient).To(BeEquivalentTo(data[1].Recipient.String()))
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
					test.NewRun(orgId()),
					test.NewRun(orgId()),
					test.NewRun(orgId()),
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
				Expect(*runs.Data[0].Id).To(BeEquivalentTo(data[1].ID.String()))
			})

			It("does not find anything if labels do not match", func() {
				runs, res := listRuns("filter[labels][abc]", "def")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(0))
			})
		})

		Describe("service", func() {
			var data dbModel.Run

			BeforeEach(func() {
				data = test.NewRun(orgId())
				Expect(db().Create(&data).Error).ToNot(HaveOccurred())
			})

			It("finds a run based on service id", func() {
				runs, res := listRuns("filter[service]", "test")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(1))
				Expect(*runs.Data[0].Id).To(BeEquivalentTo(data.ID.String()))
			})

			It("returns nothing if no such service exists", func() {
				runs, res := listRuns("filter[service]", "remediations")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Meta.Count).To(Equal(0))
			})
		})
	})

	Describe("sparse fieldsets", func() {
		BeforeEach(func() {
			run := test.NewRun(orgId())
			Expect(db().Create(&run).Error).ToNot(HaveOccurred())
		})

		DescribeTable("happy path", fieldTester(listRunsRaw),
			Entry("single field", "id"),
			Entry("defaults defined explicitly", "id", "recipient", "url", "labels", "timeout", "status"),
			Entry("all fields", "id", "recipient", "url", "labels", "timeout", "status", "created_at", "updated_at", "service", "correlation_id"),
		)

		It("400s on invalid value", func() {
			raw := listRunsRaw("fields[data]", "id,salad")
			Expect(raw.StatusCode).To(Equal(http.StatusBadRequest))
			res, err := ParseApiRunsListResponse(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.JSON400.Message).To(Equal("unknown field: salad"))
		})
	})

	Describe("sparse fieldsets satellite", func() {
		BeforeEach(func() {
			run := test.NewRun(orgId())

			playbookName := "sparse-playbookName"
			run.PlaybookName = &playbookName
			run.PlaybookRunUrl = "https://example.com/"

			Expect(db().Create(&run).Error).ToNot(HaveOccurred())
		})

		It("returns playbook run name and url", func() {
			raw := listRunsRaw("fields[data]", "name,web_console_url")
			Expect(raw.StatusCode).To(Equal(http.StatusOK))
			res, err := ParseApiRunsListResponse(raw)
			Expect(err).ToNot(HaveOccurred())

			resData := res.JSON200.Data[0]
			Expect(string(*resData.Name)).To(Equal("sparse-playbookName"))
			Expect(string(*resData.WebConsoleUrl)).To(Equal("https://example.com/"))
		})
	})

	Describe("RBAC", func() {
		var data []dbModel.Run

		BeforeEach(func() {
			data = []dbModel.Run{
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
			}

			data[0].Service = "test"
			data[1].Service = "remediations"
			data[2].Service = "salad"
			data[3].Service = "config_manager"

			Expect(db().Create(&data).Error).ToNot(HaveOccurred())
		})

		It("finds a run based on RBAC predicate", func() {
			runs, res := listRuns("fields[data]", "service")
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Meta.Count).To(Equal(3))

			expected := []interface{}{"test", "remediations", "config_manager"}
			Expect(string(*runs.Data[0].Service)).To(BeElementOf(expected...))
			Expect(string(*runs.Data[1].Service)).To(BeElementOf(expected...))
			Expect(string(*runs.Data[2].Service)).To(BeElementOf(expected...))
		})
	})

	Describe("links", func() {
		BeforeEach(func() {
			runs := []dbModel.Run{
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
			}

			Expect(db().Create(&runs).Error).ToNot(HaveOccurred())
		})

		It("returns links on no query params", func() {
			runs, res := listRuns()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))

			Expect((*runs).Links.First).To(Equal("/api/playbook-dispatcher/v1/runs?limit=50&offset=0"))
			Expect((*runs).Links.Last).To(Equal("/api/playbook-dispatcher/v1/runs?limit=50&offset=0"))
			Expect((*runs).Links.Next).To(BeNil())
			Expect((*runs).Links.Previous).To(BeNil())
		})

		It("returns links when paginating", func() {
			runs, res := listRuns("limit", 1, "offset", 1)
			Expect(res.StatusCode()).To(Equal(http.StatusOK))

			Expect((*runs).Links.First).To(Equal("/api/playbook-dispatcher/v1/runs?limit=1&offset=0"))
			Expect((*runs).Links.Last).To(Equal("/api/playbook-dispatcher/v1/runs?limit=1&offset=4"))
			Expect(*(*runs).Links.Next).To(Equal("/api/playbook-dispatcher/v1/runs?limit=1&offset=2"))
			Expect(*(*runs).Links.Previous).To(Equal("/api/playbook-dispatcher/v1/runs?limit=1&offset=0"))
		})

		It("returns links when paginating 2", func() {
			runs, res := listRuns("limit", 2, "offset", 1)
			Expect(res.StatusCode()).To(Equal(http.StatusOK))

			Expect((*runs).Links.First).To(Equal("/api/playbook-dispatcher/v1/runs?limit=2&offset=0"))
			Expect((*runs).Links.Last).To(Equal("/api/playbook-dispatcher/v1/runs?limit=2&offset=4"))
			Expect(*(*runs).Links.Next).To(Equal("/api/playbook-dispatcher/v1/runs?limit=2&offset=3"))
			Expect(*(*runs).Links.Previous).To(Equal("/api/playbook-dispatcher/v1/runs?limit=2&offset=0"))
		})

		It("propagates other query parameters", func() {
			runs, res := listRuns("limit", 1, "offset", 1, "sort_by", "created_at:desc", "fields[data]", "id", "filter[status]", "running")
			Expect(res.StatusCode()).To(Equal(http.StatusOK))

			Expect((*runs).Links.First).To(Equal("/api/playbook-dispatcher/v1/runs?fields%5Bdata%5D=id&filter%5Bstatus%5D=running&limit=1&offset=0&sort_by=created_at%3Adesc"))
			Expect((*runs).Links.Last).To(Equal("/api/playbook-dispatcher/v1/runs?fields%5Bdata%5D=id&filter%5Bstatus%5D=running&limit=1&offset=4&sort_by=created_at%3Adesc"))
			Expect(*(*runs).Links.Next).To(Equal("/api/playbook-dispatcher/v1/runs?fields%5Bdata%5D=id&filter%5Bstatus%5D=running&limit=1&offset=2&sort_by=created_at%3Adesc"))
			Expect(*(*runs).Links.Previous).To(Equal("/api/playbook-dispatcher/v1/runs?fields%5Bdata%5D=id&filter%5Bstatus%5D=running&limit=1&offset=0&sort_by=created_at%3Adesc"))
		})
	})
})
