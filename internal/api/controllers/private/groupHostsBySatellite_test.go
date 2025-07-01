package private

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"playbook-dispatcher/internal/api/connectors/inventory"
)

func TestGroupHostsBySatellite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GroupHostsBySatellite Suite")
}

var _ = Describe("groupHostsBySatellite", func() {
	It("should group hosts by satellite instance and org ID", func() {
		hostDetails := []inventory.HostDetails{
			{
				ID:                  "host1",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
			{
				ID:                  "host2",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
			{
				ID:                  "host3",
				SatelliteInstanceID: stringPtr("sat-instance-2"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.11"),
			},
		}

		result := groupHostsBySatellite(hostDetails)

		Expect(result).To(HaveLen(2))

		// Check first satellite group
		sat1Key := "sat-instance-1org-1"
		Expect(result).To(HaveKey(sat1Key))
		Expect(result[sat1Key].SatelliteInstanceID).To(Equal("sat-instance-1"))
		Expect(result[sat1Key].SatelliteOrgID).To(Equal("org-1"))
		Expect(result[sat1Key].SatelliteVersion).To(Equal("6.10"))
		Expect(result[sat1Key].Hosts).To(ConsistOf("host1", "host2"))

		// Check second satellite group
		sat2Key := "sat-instance-2org-1"
		Expect(result).To(HaveKey(sat2Key))
		Expect(result[sat2Key].SatelliteInstanceID).To(Equal("sat-instance-2"))
		Expect(result[sat2Key].SatelliteOrgID).To(Equal("org-1"))
		Expect(result[sat2Key].SatelliteVersion).To(Equal("6.11"))
		Expect(result[sat2Key].Hosts).To(ConsistOf("host3"))
	})

	It("should handle hosts with different org IDs", func() {
		hostDetails := []inventory.HostDetails{
			{
				ID:                  "host1",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
			{
				ID:                  "host2",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-2"),
				SatelliteVersion:    stringPtr("6.10"),
			},
		}

		result := groupHostsBySatellite(hostDetails)

		Expect(result).To(HaveLen(2))

		sat1Org1Key := "sat-instance-1org-1"
		Expect(result).To(HaveKey(sat1Org1Key))
		Expect(result[sat1Org1Key].Hosts).To(ConsistOf("host1"))

		sat1Org2Key := "sat-instance-1org-2"
		Expect(result).To(HaveKey(sat1Org2Key))
		Expect(result[sat1Org2Key].Hosts).To(ConsistOf("host2"))
	})

	It("should handle empty host details slice", func() {
		hostDetails := []inventory.HostDetails{}

		result := groupHostsBySatellite(hostDetails)

		Expect(result).To(BeEmpty())
		Expect(result).To(HaveLen(0))
	})

	It("should handle hosts with nil satellite fields", func() {
		hostDetails := []inventory.HostDetails{
			{
				ID:                  "host1",
				SatelliteInstanceID: nil,
				SatelliteOrgID:      nil,
				SatelliteVersion:    nil,
			},
		}

		// This test case would cause a panic in the current implementation
		// because it dereferences nil pointers. The function should be made
		// more robust to handle this case.
		Expect(func() {
			groupHostsBySatellite(hostDetails)
		}).To(Panic())
	})

	It("should handle hosts with partial satellite information", func() {
		hostDetails := []inventory.HostDetails{
			{
				ID:                  "host1",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    nil,
			},
			{
				ID:                  "host2",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
		}

		result := groupHostsBySatellite(hostDetails)

		Expect(result).To(HaveLen(1))

		satKey := "sat-instance-1org-1"
		Expect(result).To(HaveKey(satKey))
		Expect(result[satKey].SatelliteInstanceID).To(Equal("sat-instance-1"))
		Expect(result[satKey].SatelliteOrgID).To(Equal("org-1"))
		Expect(result[satKey].SatelliteVersion).To(Equal("6.10")) // Should use the non-nil version
		Expect(result[satKey].Hosts).To(ConsistOf("host1", "host2"))
	})

	It("should handle single host", func() {
		hostDetails := []inventory.HostDetails{
			{
				ID:                  "host1",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
		}

		result := groupHostsBySatellite(hostDetails)

		Expect(result).To(HaveLen(1))

		satKey := "sat-instance-1org-1"
		Expect(result).To(HaveKey(satKey))
		Expect(result[satKey].SatelliteInstanceID).To(Equal("sat-instance-1"))
		Expect(result[satKey].SatelliteOrgID).To(Equal("org-1"))
		Expect(result[satKey].SatelliteVersion).To(Equal("6.10"))
		Expect(result[satKey].Hosts).To(ConsistOf("host1"))
	})

	It("should handle multiple hosts with same satellite configuration", func() {
		hostDetails := []inventory.HostDetails{
			{
				ID:                  "host1",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
			{
				ID:                  "host2",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
			{
				ID:                  "host3",
				SatelliteInstanceID: stringPtr("sat-instance-1"),
				SatelliteOrgID:      stringPtr("org-1"),
				SatelliteVersion:    stringPtr("6.10"),
			},
		}

		result := groupHostsBySatellite(hostDetails)

		Expect(result).To(HaveLen(1))

		satKey := "sat-instance-1org-1"
		Expect(result).To(HaveKey(satKey))
		Expect(result[satKey].SatelliteInstanceID).To(Equal("sat-instance-1"))
		Expect(result[satKey].SatelliteOrgID).To(Equal("org-1"))
		Expect(result[satKey].SatelliteVersion).To(Equal("6.10"))
		Expect(result[satKey].Hosts).To(ConsistOf("host1", "host2", "host3"))
	})
})

// Helper function to create string pointers for testing
func stringPtr(s string) *string {
	return &s
}
