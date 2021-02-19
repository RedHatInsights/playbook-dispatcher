package common

import (
	"context"
	"net/http"
	"playbook-dispatcher/internal/common/utils/test"
)

const TestServer = "http://localhost:9002"

var (
	TestClient        = &test.Client
	TestRequestEditor = func(ctx context.Context, req *http.Request) error {
		if account := ctx.Value(accountContextKey); account != nil {
			req.Header.Set("x-rh-identity", test.IdentityHeaderMinimal(account.(string)))
		}

		return nil
	}
)

const accountContextKey = iota

func ContextWithIdentity(account string) context.Context {
	return context.WithValue(test.TestContext(), accountContextKey, account)
}
