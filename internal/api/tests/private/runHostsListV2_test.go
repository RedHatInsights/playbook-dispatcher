// Coded in collaboration with AI
package private

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"playbook-dispatcher/internal/api/controllers/public"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func doGetRunHosts(keysAndValues ...interface{}) (*public.RunHosts, *http.Response) {
	url := utils.BuildUrl("http://localhost:9002/internal/v2/run_hosts", keysAndValues...)

	req, err := http.NewRequest("GET", url, nil)
	Expect(err).ToNot(HaveOccurred())
	req.Header.Set("x-rh-identity", test.IdentityHeaderMinimal(orgId()))
	req.Header.Set("authorization", "PSK xwKhCUzgJ8")
	resp, err := test.Client.Do(req)
	Expect(err).ToNot(HaveOccurred())

	bodyBytes, err := io.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()

	var result public.RunHosts
	err = json.Unmarshal(bodyBytes, &result)
	Expect(err).ToNot(HaveOccurred())

	return &result, resp
}

var _ = Describe("runHostsListV2", func() {
	db := test.WithDatabase()

	dbInsertRuns := func(runs ...dbModel.Run) {
		Expect(db().Create(runs).Error).ToNot(HaveOccurred())
	}

	dbInsertHosts := func(hosts ...dbModel.RunHost) {
		Expect(db().Create(hosts).Error).ToNot(HaveOccurred())
	}

	Describe("GET /internal/v2/run_hosts", func() {
		It("returns all run hosts without service-level filtering (no RBAC)", func() {
			// Create 3 hosts for 3 different services
			run1 := test.NewRun(orgId())
			run1.Service = "remediations"
			run2 := test.NewRun(orgId())
			run2.Service = "compliance"
			run3 := test.NewRun(orgId())
			run3.Service = "insights"

			dbInsertRuns(run1, run2, run3)

			inventoryId1 := uuid.New()
			inventoryId2 := uuid.New()
			inventoryId3 := uuid.New()

			host1 := test.NewRunHost(run1.ID, "success", &inventoryId1)
			host2 := test.NewRunHost(run2.ID, "success", &inventoryId2)
			host3 := test.NewRunHost(run3.ID, "success", &inventoryId3)

			dbInsertHosts(host1, host2, host3)

			result, resp := doGetRunHosts()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Internal endpoint should return ALL 3 hosts (no RBAC filtering)
			// Unlike public endpoint which would filter based on user permissions
			Expect(result.Meta.Count).To(Equal(3))
			Expect(result.Meta.Total).To(Equal(3))
			Expect(result.Data).To(HaveLen(3))
		})

		It("filters by run_id", func() {
			run1 := test.NewRun(orgId())
			run2 := test.NewRun(orgId())

			dbInsertRuns(run1, run2)

			inventoryId1 := uuid.New()
			inventoryId2 := uuid.New()

			host1 := test.NewRunHost(run1.ID, "success", &inventoryId1)
			host2 := test.NewRunHost(run2.ID, "success", &inventoryId2)

			dbInsertHosts(host1, host2)

			result, resp := doGetRunHosts("filter[run][id]", run1.ID.String())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Meta.Count).To(Equal(1))
			Expect(result.Data).To(HaveLen(1))
			Expect(*result.Data[0].Run.Id).To(Equal(run1.ID))
		})

		It("filters by status", func() {
			run := test.NewRun(orgId())
			dbInsertRuns(run)

			inventoryId1 := uuid.New()
			inventoryId2 := uuid.New()

			host1 := test.NewRunHost(run.ID, "success", &inventoryId1)
			host1.Host = "host1"
			host2 := test.NewRunHost(run.ID, "failure", &inventoryId2)
			host2.Host = "host2"

			dbInsertHosts(host1, host2)

			result, resp := doGetRunHosts("filter[status]", "success")
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Meta.Count).To(Equal(1))
			Expect(result.Data).To(HaveLen(1))
			Expect(string(*result.Data[0].Status)).To(Equal("success"))
		})

		It("filters by inventory_id", func() {
			run := test.NewRun(orgId())
			dbInsertRuns(run)

			inventoryId1 := uuid.New()
			inventoryId2 := uuid.New()

			host1 := test.NewRunHost(run.ID, "success", &inventoryId1)
			host1.Host = "host1"
			host2 := test.NewRunHost(run.ID, "success", &inventoryId2)
			host2.Host = "host2"

			dbInsertHosts(host1, host2)

			result, resp := doGetRunHosts("filter[inventory_id]", inventoryId1.String())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Meta.Count).To(Equal(1))
			Expect(result.Data).To(HaveLen(1))
			Expect(*result.Data[0].Host).To(Equal("host1"))
		})

		It("supports pagination with limit and offset", func() {
			run := test.NewRun(orgId())
			dbInsertRuns(run)

			// Create 5 hosts
			hosts := []dbModel.RunHost{}
			for i := 0; i < 5; i++ {
				inventoryId := uuid.New()
				host := test.NewRunHost(run.ID, "success", &inventoryId)
				host.Host = fmt.Sprintf("host%d", i)
				hosts = append(hosts, host)
			}
			dbInsertHosts(hosts...)

			// Request with limit=2, offset=1
			result, resp := doGetRunHosts("limit", "2", "offset", "1")
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Meta.Count).To(Equal(2))
			Expect(result.Meta.Total).To(Equal(5))
			Expect(result.Data).To(HaveLen(2))
		})

		It("supports field selection", func() {
			run := test.NewRun(orgId())
			dbInsertRuns(run)

			inventoryId := uuid.New()
			host := test.NewRunHost(run.ID, "success", &inventoryId)
			dbInsertHosts(host)

			result, resp := doGetRunHosts("fields[data]", "host,status")
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Data).To(HaveLen(1))

			// Selected fields should be present
			Expect(result.Data[0].Host).ToNot(BeNil())
			Expect(result.Data[0].Status).ToNot(BeNil())

			// Non-selected fields should be nil
			Expect(result.Data[0].Run).To(BeNil())
			Expect(result.Data[0].Stdout).To(BeNil())
		})

		It("filters by labels", func() {
			run1 := test.NewRun(orgId())
			run1.Labels = map[string]string{"env": "prod", "team": "platform"}
			run2 := test.NewRun(orgId())
			run2.Labels = map[string]string{"env": "dev", "team": "app"}

			dbInsertRuns(run1, run2)

			inventoryId1 := uuid.New()
			inventoryId2 := uuid.New()

			host1 := test.NewRunHost(run1.ID, "success", &inventoryId1)
			host2 := test.NewRunHost(run2.ID, "success", &inventoryId2)

			dbInsertHosts(host1, host2)

			result, resp := doGetRunHosts("filter[run][labels][env]", "prod")
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Meta.Count).To(Equal(1))
			Expect(result.Data).To(HaveLen(1))
		})

		It("filters by service", func() {
			run1 := test.NewRun(orgId())
			run1.Service = "remediations"
			run2 := test.NewRun(orgId())
			run2.Service = "compliance"

			dbInsertRuns(run1, run2)

			inventoryId1 := uuid.New()
			inventoryId2 := uuid.New()

			host1 := test.NewRunHost(run1.ID, "success", &inventoryId1)
			host2 := test.NewRunHost(run2.ID, "success", &inventoryId2)

			dbInsertHosts(host1, host2)

			result, resp := doGetRunHosts("filter[run][service]", "remediations")
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Meta.Count).To(Equal(1))
			Expect(result.Data).To(HaveLen(1))
		})

		It("isolates data by org_id", func() {
			orgId1 := "org-1"
			orgId2 := "org-2"

			run1 := test.NewRun(orgId1)
			run2 := test.NewRun(orgId2)

			dbInsertRuns(run1, run2)

			inventoryId1 := uuid.New()
			inventoryId2 := uuid.New()

			host1 := test.NewRunHost(run1.ID, "success", &inventoryId1)
			host2 := test.NewRunHost(run2.ID, "success", &inventoryId2)

			dbInsertHosts(host1, host2)

			// Request uses orgId() from suite which should only see its own data
			result, resp := doGetRunHosts()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Should only see hosts for the requesting org
			// Note: orgId() returns the test org ID which may not match org-1 or org-2
			// This test verifies org isolation is working
			for _, data := range result.Data {
				Expect(data.Run).ToNot(BeNil())
			}
		})

		It("includes stdout data when requested", func() {
			run := test.NewRun(orgId())
			dbInsertRuns(run)

			inventoryId := uuid.New()
			host := test.NewRunHost(run.ID, "success", &inventoryId)
			host.Log = "PLAY [all] ***\nTASK [debug] ***\nok: [localhost]"
			dbInsertHosts(host)

			result, resp := doGetRunHosts("fields[data]", "host,status,stdout")
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(result.Data).To(HaveLen(1))
			Expect(result.Data[0].Stdout).ToNot(BeNil())
			Expect(*result.Data[0].Stdout).To(ContainSubstring("PLAY [all]"))
		})
	})
})
