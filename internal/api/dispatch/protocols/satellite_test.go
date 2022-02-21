package protocols

import (
	"playbook-dispatcher/internal/common/model/generic"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Satellite Protocol", func() {
	It("uses the correct directive", func() {
		Expect(string(SatelliteProtocol.GetDirective())).To(Equal("rhc-cloud-connector-worker"))
	})

	Describe("metadata", func() {
		It("produces correct metadata", func() {
			satID := uuid.New()

			run := generic.RunInput{
				Name:          utils.StringRef("Red Hat Playbook"),
				WebConsoleUrl: utils.StringRef("https://console.redhat.com/insights/remediations"),
				Principal:     utils.StringRef("jharting"),
				SatId:         &satID,
				SatOrgId:      utils.StringRef("1"),
				Hosts: []generic.RunHostsInput{
					{InventoryId: utils.UUIDRef(uuid.MustParse("ee4bbcd0-a782-4335-a904-c690b60ec4c4"))},
					{InventoryId: utils.UUIDRef(uuid.MustParse("330d5e16-7110-4fe1-a40d-cbb084e50aae"))},
					{InventoryId: utils.UUIDRef(uuid.MustParse("304f9557-420c-4009-97ae-4dafbf5c8dc5"))},
				},
			}

			correlationID := uuid.New()

			cfg := viper.New()
			cfg.Set("response.interval", "3")
			cfg.Set("return.url", "https://example.com")
			cfg.Set("satellite.response.full", true)

			metadata := SatelliteProtocol.BuildMedatada(run, correlationID, cfg)
			Expect(metadata).To(HaveLen(11))
			Expect(metadata["operation"]).To(Equal("run"))
			Expect(metadata["return_url"]).To(Equal("https://example.com"))
			Expect(metadata["correlation_id"]).To(Equal(correlationID.String()))
			Expect(metadata["playbook_run_name"]).To(Equal(*run.Name))
			Expect(metadata["playbook_run_url"]).To(Equal(*run.WebConsoleUrl))
			Expect(metadata["sat_id"]).To(Equal(satID.String()))
			Expect(metadata["sat_org_id"]).To(Equal(*run.SatOrgId))
			Expect(metadata["initiator_user_id"]).To(Equal("6874f78b19f011c53313a129b5ef6d6991456c0dcee89c8882c862ab7e8ea6dd"))
			Expect(metadata["hosts"]).To(Equal("ee4bbcd0-a782-4335-a904-c690b60ec4c4,330d5e16-7110-4fe1-a40d-cbb084e50aae,304f9557-420c-4009-97ae-4dafbf5c8dc5"))
			Expect(metadata["response_interval"]).To(Equal("3"))
			Expect(metadata["response_full"]).To(Equal("true"))
		})
	})
})
