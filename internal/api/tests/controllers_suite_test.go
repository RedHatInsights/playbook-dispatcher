package tests

import (
	"context"
	"net/http"
	"playbook-dispatcher/internal/api"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils/test"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg := config.Get()
	cfg.Set("web.port", 9002)

	api.WithApi(cfg)

	RunSpecs(t, "Controllers Suite")
}

var (
	accountNumber = test.WithAccountNumber()

	client = &Client{
		Server: "http://localhost:9002",
		Client: &test.Client,
		RequestEditor: func(ctx context.Context, req *http.Request) error {
			if account := ctx.Value(accountContextKey); account != nil {
				req.Header.Set("x-rh-identity", test.IdentityHeaderMinimal(account.(string)))
			}

			return nil
		},
	}
)

const accountContextKey = iota

func ContextWithIdentity(account string) context.Context {
	return context.WithValue(test.TestContext(), accountContextKey, account)
}
