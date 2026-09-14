package responseConsumer

import (
	"playbook-dispatcher/internal/common/utils/test"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Response Consumer Suite")
}

var (
	orgId = test.WithOrgId()
)
