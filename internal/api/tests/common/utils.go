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
		if org_id := ctx.Value(contextKey); org_id != nil {
			req.Header.Set("x-rh-identity", test.IdentityHeaderMinimal(org_id.(string)))
		}

		return nil
	}
)

type key int

const contextKey key = iota

func ContextWithIdentity(org_id string) context.Context {
	return context.WithValue(test.TestContext(), contextKey, org_id)
}
