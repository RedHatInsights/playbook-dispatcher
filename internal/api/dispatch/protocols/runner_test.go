package protocols

import (
	"playbook-dispatcher/internal/common/model/generic"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Runner Protocol", func() {
	It("uses the correct directive", func() {
		Expect(string(RunnerProtocol.GetDirective())).To(Equal("rhc-worker-playbook"))
	})

	Describe("metadata", func() {
		It("produces correct metadata", func() {
			run := generic.RunInput{}

			correlationID := uuid.New()
			cfg := viper.New()
			cfg.Set("response.interval", "3")
			cfg.Set("return.url", "https://example.com")

			metadata := RunnerProtocol.BuildMetaData(run, correlationID, cfg)
			Expect(metadata).To(HaveLen(3))
			Expect(metadata["crc_dispatcher_correlation_id"]).To(Equal(correlationID.String()))
			Expect(metadata["response_interval"]).To(Equal("3"))
			Expect(metadata["return_url"]).To(Equal("https://example.com"))
		})
	})
})
