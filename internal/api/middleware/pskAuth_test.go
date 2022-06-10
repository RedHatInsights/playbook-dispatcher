package middleware

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

const key = "secret"

func testPskAuth(req *http.Request) (*httptest.ResponseRecorder, error) {
	recorder := httptest.NewRecorder()

	config := map[string]string{
		"principal1": key,
	}

	handler := CheckPskAuth(config)(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, GetPSKPrincipal(ctx.Request().Context()))
	})

	c := echo.New().NewContext(req, recorder)

	return recorder, handler(c)
}

func newReqInternal() *http.Request {
	req, err := http.NewRequest("GET", "/internal/dispatch", nil)
	Expect(err).ToNot(HaveOccurred())
	req = req.WithContext(utils.SetLog(context.Background(), zap.NewNop().Sugar()))
	return req
}

var _ = Describe("PSK auth middleware", func() {
	It("401s on no auth header", func() {
		_, err := testPskAuth(newReqInternal())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("code=401, message=Unauthorized"))
	})

	It("401s on unsupported auth scheme", func() {
		req := newReqInternal()
		req.Header.Set("authorization", "Basic Zm9vOmJhcg==")
		_, err := testPskAuth(req)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("code=401, message=Unsupported authentication key format"))
	})

	It("403s on unknown key", func() {
		req := newReqInternal()
		req.Header.Set("authorization", "PSK foobar")
		_, err := testPskAuth(req)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("code=403, message=Forbidden"))
	})

	It("passthrough on known key", func() {
		req := newReqInternal()
		req.Header.Set("authorization", fmt.Sprintf("PSK %s", key))
		res, err := testPskAuth(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Result().StatusCode).To(Equal(200))
	})

	It("makes the principal available in context", func() {
		req := newReqInternal()
		req.Header.Set("authorization", fmt.Sprintf("PSK %s", key))
		res, err := testPskAuth(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Result().StatusCode).To(Equal(200))
		body, err := ioutil.ReadAll(res.Result().Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(BeEquivalentTo("principal1"))
	})
})
