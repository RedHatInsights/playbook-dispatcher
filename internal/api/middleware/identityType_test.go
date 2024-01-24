package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	identityMw "github.com/redhatinsights/platform-go-middlewares/identity"
)

var _ = Describe("Identity type middleware", func() {
	var req *http.Request

	BeforeEach(func() {
		var err error
		req, err = http.NewRequest("GET", "/api/playbook-dispatcher/v1/runs", nil)
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("Only accepts User identity",
		func(identityType string, expectedStatus int, expectedBody string) {
			identity := fmt.Sprintf(`{ "identity": {"account_number": "540155", "type": "%s", "internal": { "org_id": "1979710" } } }`, identityType)
			req.Header.Set("x-rh-identity", base64.StdEncoding.EncodeToString([]byte(identity)))
			recorder := httptest.NewRecorder()
			handler := identityMw.EnforceIdentity(EnforceIdentityType(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})))
			handler.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(expectedStatus))
			Expect(recorder.Body.String()).To(Equal(expectedBody))
		},

		Entry("User", "User", 200, ""),
		Entry("ServiceAccount", "ServiceAccount", 200, ""),
		Entry("System", "System", 403, "unauthorized principal type: system\n"),
		Entry("Random", "salad", 403, "unauthorized principal type: salad\n"),
	)
})
