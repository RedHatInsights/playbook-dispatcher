package public

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Middleware", func() {
	Describe("request id", func() {
		const requestIdHeader = "x-rh-insights-request-id"

		It("attaches request id to the response", func() {
			const requestId = "33ee136c-da81-4a68-953c-c22bdd096d30"

			req, err := http.NewRequest(http.MethodGet, "http://localhost:9002/api/playbook-dispatcher/v1/runs", nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add(requestIdHeader, requestId)

			res, err := test.Client.Do(req)
			Expect(err).To(HaveOccurred())
			Expect(res.Header.Get(requestIdHeader)).To(Equal(requestId))
		})
	})

	Describe("api spec", func() {
		It("openapi.json can be downloaded", func() {
			req, err := http.NewRequest(http.MethodGet, "http://localhost:9002/api/playbook-dispatcher/v1/openapi.json", nil)
			Expect(err).ToNot(HaveOccurred())
			res, err := test.Client.Do(req)
			Expect(err).To(HaveOccurred())

			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("identity", func() {
		const identityHeader = "x-rh-identity"

		It("identity header enforced on public route", func() {
			req, err := http.NewRequest(http.MethodGet, "http://localhost:9002/api/playbook-dispatcher/v1/runs", nil)
			req.Header.Add(identityHeader, "eyJpZGVudGl0eSI6eyJpbnRlcm5hbCI6eyJvcmdfaWQiOiI1MzE4MjkwIn0sImFjY291bnRfbnVtYmVyIjoiOTAxNTc4IiwidXNlciI6e30sInR5cGUiOiJVc2VyIn19Cg==")
			Expect(err).ToNot(HaveOccurred())
			res, err := test.Client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})

		It("identity header enforced on public route (negative)", func() {
			req, err := http.NewRequest(http.MethodGet, "http://localhost:9002/api/playbook-dispatcher/v1/runs", nil)
			Expect(err).ToNot(HaveOccurred())
			res, err := test.Client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
			data, _ := ioutil.ReadAll(res.Body)
			defer res.Body.Close()

			Expect(data).To(BeEquivalentTo("Bad Request: missing x-rh-identity header\n"))
		})
	})

	Describe("openapi validator", func() {
		It("rejects invalid request", func() {
			req, err := http.NewRequest(http.MethodPost, "http://localhost:9002/internal/dispatch", bytes.NewBuffer([]byte("[]")))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			req.Header.Add("authorization", "PSK xwKhCUzgJ8")
			res, err := test.Client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.StatusCode).To(Equal(http.StatusBadRequest))

			var parsed map[string]string
			Expect(json.NewDecoder(res.Body).Decode(&parsed)).ToNot(HaveOccurred())

			Expect(parsed["message"]).To(Equal("Request body has an error: doesn't match the schema: Minimum number of items is 1"))
		})
	})
})
