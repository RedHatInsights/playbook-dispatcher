package private

import (
	"bytes"
	"io"
	"net/http"

	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	Describe("get internal version", func() {
		It("should return the git revision the API was built from", func() {

			req, err := http.NewRequest(http.MethodGet, "http://localhost:9002/internal/version", nil)
			Expect(err).ToNot(HaveOccurred())

			res, err := test.Client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(http.StatusOK))

			data, err := io.ReadAll(res.Body)
			Expect(err).ToNot(HaveOccurred())

			// Remove quotes and newline from the returned data
			data = bytes.Replace(data, []byte("\""), []byte(""), 2)
			data = bytes.Replace(data, []byte("\n"), []byte(""), 1)

			Expect(string(data)).To(BeEquivalentTo(testBuildCommit))

		})
	})
})
