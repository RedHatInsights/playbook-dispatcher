package public

import (
	"playbook-dispatcher/internal/api"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils/test"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	accountNumber = test.WithAccountNumber()
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg := config.Get()
	cfg.Set("web.port", 9002)

	api.WithApi(cfg)

	RunSpecs(t, "Public Controller Suite")
}
