package private

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api"
	"playbook-dispatcher/internal/api/tests/common"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils/test"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	pskKey = "pskKey"

	webConsoleUrlDefault = "https://example.com"
	testBuildCommit      = "testV1"
)

var (
	orgId         = test.WithOrgId()
	accountNumber = test.WithAccountNumber()
	client        = &Client{
		Server: common.TestServer,
		Client: common.TestClient,
		RequestEditors: []RequestEditorFn{func(ctx context.Context, req *http.Request) error {
			if psk, ok := ctx.Value(pskKey).(string); ok {
				req.Header.Set("authorization", fmt.Sprintf("PSK %s", psk))
			} else {
				req.Header.Set("authorization", "PSK xwKhCUzgJ8")
			}

			return nil
		}},
	}
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg := config.Get()
	cfg.Set("web.port", 9002)
	cfg.Set("cloud.connector.rps", 5)
	cfg.Set("cloud.connector.req.bucket", 5)

	cfg.Set("build.commit", testBuildCommit)

	cfg.Set("web.console.url.default", webConsoleUrlDefault)

	api.WithApi(cfg)

	RunSpecs(t, "Private Controller Suite")
}
