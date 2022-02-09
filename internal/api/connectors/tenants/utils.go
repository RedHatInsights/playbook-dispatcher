package tenants

import (
	"net/http"
	"playbook-dispatcher/internal/common/utils"
)

type operationAwareDoer interface {
	Do(req *http.Request, operation string) (*http.Response, error)
	Unwrap() utils.HttpRequestDoer
}

type operationAwareDoerImpl struct {
	client utils.HttpRequestDoer
}

func (this *operationAwareDoerImpl) Do(req *http.Request, operation string) (*http.Response, error) {
	return this.client.Do(req)
}

func (this *operationAwareDoerImpl) Unwrap() utils.HttpRequestDoer {
	return this.client
}
