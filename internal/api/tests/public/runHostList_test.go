package public

import (
	"net/http"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils/test"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func listRunHosts(keysAndValues ...interface{}) (*RunHosts, *ApiRunHostsListResponse) {
	raw := listRunHostsRaw(keysAndValues...)
	res, err := ParseApiRunHostsListResponse(raw)
	Expect(err).ToNot(HaveOccurred())

	return res.JSON200, res
}

func listRunHostsRaw(keysAndValues ...interface{}) *http.Response {
	return doGet("http://localhost:9002/api/playbook-dispatcher/v1/run_hosts", keysAndValues...)
}

var _ = Describe("runHostList", func() {
	db := test.WithDatabase()

	dbInsertRuns := func(runs ...dbModel.Run) {
		Expect(db().Create(runs).Error).ToNot(HaveOccurred())
	}

	dbInsertHosts := func(hosts ...dbModel.RunHost) {
		Expect(db().Create(hosts).Error).ToNot(HaveOccurred())
	}

	Describe("list hosts", func() {
		It("by default returns a list of run hosts", func() {
			run := test.NewRun(orgId())
			dbInsertRuns(run)
			host1, host2 := test.NewRunHost(run.ID, "running", nil), test.NewRunHost(run.ID, "failure", nil)
			host1.Host = "01.example.com"
			host2.Host = "02.example.com"
			dbInsertHosts(host1, host2)

			runs, res := listRunHosts()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Data).To(HaveLen(2))
			Expect([]RunStatus{*runs.Data[0].Status, *runs.Data[1].Status}).To(ContainElements(RunStatusFailure, RunStatusRunning))
			Expect([]string{*runs.Data[0].Host, *runs.Data[1].Host}).To(ContainElements(host1.Host, host2.Host))
			Expect(*runs.Data[0].Run.Id).To(BeEquivalentTo(run.ID))
			Expect(*runs.Data[1].Run.Id).To(BeEquivalentTo(run.ID))
		})

		Describe("filtering", func() {
			It("filters by host status", func() {
				data := []dbModel.Run{
					test.NewRunWithStatus(orgId(), "success"),
					test.NewRunWithStatus(orgId(), "failure"),
				}

				dbInsertRuns(data...)
				dbInsertHosts(test.NewRunHost(data[0].ID, "success", nil), test.NewRunHost(data[1].ID, "failure", nil))

				runs, res := listRunHosts("filter[status]", "failure")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(1))
				Expect(*runs.Data[0].Run.Id).To(BeEquivalentTo(data[1].ID))
			})

			It("filters by run id", func() {
				data := []dbModel.Run{
					test.NewRun(orgId()),
					test.NewRun(orgId()),
					test.NewRun(orgId()),
				}

				dbInsertRuns(data...)
				dbInsertHosts(test.MapRunToHost(data, func(run dbModel.Run) dbModel.RunHost {
					return test.NewRunHost(run.ID, "running", nil)
				})...)

				runs, res := listRunHosts("filter[run][id]", data[1].ID.String())
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(1))
				Expect(*runs.Data[0].Run.Id).To(BeEquivalentTo(data[1].ID))
			})

			It("filters by run labels", func() {
				data := []dbModel.Run{
					test.NewRun(orgId()),
					test.NewRun(orgId()),
					test.NewRun(orgId()),
				}

				data[0].Labels = map[string]string{"remediation": "0"}
				data[1].Labels = map[string]string{"remediation": "1"}
				data[2].Labels = map[string]string{"remediation": "2"}

				dbInsertRuns(data...)
				dbInsertHosts(test.MapRunToHost(data, func(run dbModel.Run) dbModel.RunHost {
					return test.NewRunHost(run.ID, "running", nil)
				})...)

				runs, res := listRunHosts("filter[run][labels][remediation]", "2")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(1))
				Expect(*runs.Data[0].Run.Id).To(BeEquivalentTo(data[2].ID))
			})

			It("filters by service", func() {
				run := test.NewRun(orgId())
				dbInsertRuns(run)
				dbInsertHosts(test.NewRunHost(run.ID, "running", nil))

				runs, res := listRunHosts("filter[run][service]", "test")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(1))
				Expect(*runs.Data[0].Run.Id).To(BeEquivalentTo(run.ID))

				runs, res = listRunHosts("filter[run][service]", "remediations")
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(0))
			})

			It("filters by inventory_id", func() {
				data := []dbModel.Run{
					test.NewRun(orgId()),
					test.NewRun(orgId()),
					test.NewRun(orgId()),
				}

				dbInsertRuns(data...)
				hosts := test.MapRunToHost(data, func(run dbModel.Run) dbModel.RunHost {
					inventoryID := uuid.New()
					return test.NewRunHost(run.ID, "running", &inventoryID)
				})

				dbInsertHosts(hosts...)

				runs, res := listRunHosts("filter[inventory_id]", hosts[1].InventoryID.String())
				Expect(res.StatusCode()).To(Equal(http.StatusOK))
				Expect(runs.Data).To(HaveLen(1))
				Expect(*runs.Data[0].Run.Id).To(BeEquivalentTo(data[1].ID))
			})

			It("handle invalid inventory id filter", func() {
				_, res := listRunHosts("filter[inventory_id]", "fred-flintstone-barney-rubble-not-uuid")
				Expect(res.StatusCode()).To(Equal(http.StatusBadRequest))
				Expect(res.JSON400.Message).To(Equal("Unable to parse inventory id!"))
			})
		})
	})

	Describe("sparse fieldsets", func() {
		BeforeEach(func() {
			run := test.NewRun(orgId())
			dbInsertRuns(run)
			inventoryID := uuid.New()
			dbInsertHosts(test.NewRunHost(run.ID, "running", &inventoryID))
		})

		DescribeTable("happy path", fieldTester(listRunHostsRaw),
			Entry("single field", "host"),
			Entry("defaults defined explicitly", "host", "status", "run"),
			Entry("all fields", "host", "status", "run", "stdout", "links", "inventory_id"),
		)

		It("400s on invalid value", func() {
			raw := listRunHostsRaw("fields[data]", "host,salad")
			Expect(raw.StatusCode).To(Equal(http.StatusBadRequest))
			res, err := ParseApiRunHostsListResponse(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.JSON400.Message).To(Equal("unknown field: salad"))
		})
	})

	DescribeTable("pagination",
		func(expected, limit, offset int) {
			newRuns := test.NewRunsWithLocalhost(orgId(), 3)
			dbInsertRuns(newRuns...)

			dbInsertHosts(test.FlatMapRunToHost(newRuns, func(run dbModel.Run) []dbModel.RunHost {
				return []dbModel.RunHost{
					test.NewRunHostWithHostname(run.ID, "running", "host1"),
					test.NewRunHostWithHostname(run.ID, "running", "host2"),
					test.NewRunHostWithHostname(run.ID, "running", "host3"),
					test.NewRunHostWithHostname(run.ID, "running", "host4"),
				}
			})...)

			runs, res := listRunHosts("limit", limit, "offset", offset)
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Data).To(HaveLen(expected))
			Expect(runs.Meta.Count).To(Equal(expected))
			Expect(runs.Meta.Total).To(Equal(12))
		},

		Entry("limit=2", 2, 2, 0),
		Entry("limit=5", 5, 5, 0),
		Entry("limit=5, offset=10", 2, 5, 10),
		Entry("limit=5, offset=20", 0, 5, 20),
	)

	Describe("RBAC", func() {
		var data []dbModel.Run

		BeforeEach(func() {
			data = []dbModel.Run{
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
			}

			data[0].Service = "test"
			data[1].Service = "remediations"
			data[2].Service = "salad"

			dbInsertRuns(data...)
			dbInsertHosts(test.MapRunToHost(data, func(run dbModel.Run) dbModel.RunHost {
				return test.NewRunHost(run.ID, "running", nil)
			})...)
		})

		It("finds a run based on RBAC predicate", func() {
			runs, res := listRunHosts()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))
			Expect(runs.Meta.Count).To(Equal(2))

			expected := []uuid.UUID{data[0].ID, data[1].ID}
			Expect(*runs.Data[0].Run.Id).To(BeElementOf(expected))
			Expect(*runs.Data[1].Run.Id).To(BeElementOf(expected))
		})
	})

	Describe("links", func() {
		BeforeEach(func() {
			var data = []dbModel.Run{
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
				test.NewRun(orgId()),
			}

			dbInsertRuns(data...)
			dbInsertHosts(test.MapRunToHost(data, func(run dbModel.Run) dbModel.RunHost {
				return test.NewRunHost(run.ID, "running", nil)
			})...)
		})

		It("returns links on no query params", func() {
			runs, res := listRunHosts()
			Expect(res.StatusCode()).To(Equal(http.StatusOK))

			Expect((*runs).Links.First).To(Equal("/api/playbook-dispatcher/v1/run_hosts?limit=50&offset=0"))
			Expect((*runs).Links.Last).To(Equal("/api/playbook-dispatcher/v1/run_hosts?limit=50&offset=0"))
			Expect((*runs).Links.Next).To(BeNil())
			Expect((*runs).Links.Previous).To(BeNil())
		})

		It("returns links when paginating", func() {
			runs, res := listRunHosts("limit", 1, "offset", 1)
			Expect(res.StatusCode()).To(Equal(http.StatusOK))

			Expect((*runs).Links.First).To(Equal("/api/playbook-dispatcher/v1/run_hosts?limit=1&offset=0"))
			Expect((*runs).Links.Last).To(Equal("/api/playbook-dispatcher/v1/run_hosts?limit=1&offset=4"))
			Expect(*(*runs).Links.Next).To(Equal("/api/playbook-dispatcher/v1/run_hosts?limit=1&offset=2"))
			Expect(*(*runs).Links.Previous).To(Equal("/api/playbook-dispatcher/v1/run_hosts?limit=1&offset=0"))
		})

		It("propagates other query parameters", func() {
			runs, res := listRunHosts("limit", 1, "offset", 1, "fields[data]", "host", "filter[status]", "running")
			Expect(res.StatusCode()).To(Equal(http.StatusOK))

			Expect((*runs).Links.First).To(Equal("/api/playbook-dispatcher/v1/run_hosts?fields%5Bdata%5D=host&filter%5Bstatus%5D=running&limit=1&offset=0"))
			Expect((*runs).Links.Last).To(Equal("/api/playbook-dispatcher/v1/run_hosts?fields%5Bdata%5D=host&filter%5Bstatus%5D=running&limit=1&offset=4"))
			Expect(*(*runs).Links.Next).To(Equal("/api/playbook-dispatcher/v1/run_hosts?fields%5Bdata%5D=host&filter%5Bstatus%5D=running&limit=1&offset=2"))
			Expect(*(*runs).Links.Previous).To(Equal("/api/playbook-dispatcher/v1/run_hosts?fields%5Bdata%5D=host&filter%5Bstatus%5D=running&limit=1&offset=0"))
		})
	})
})
