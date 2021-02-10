package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

type mockHttpRequestDoer struct {
	request  *http.Request
	response *http.Response
}

func (this *mockHttpRequestDoer) Do(req *http.Request) (*http.Response, error) {
	this.request = req
	return this.response, nil
}

func withMockResponse(statusCode int, body string) mockHttpRequestDoer {
	response := http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(body))),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	return mockHttpRequestDoer{
		response: &response,
	}
}

var _ = Describe("Cloud Connector", func() {
	It("interprets the response correctly", func() {
		doer := withMockResponse(200, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), zap.NewNop().Sugar(), &doer)
		ctx := utils.SetLog(context.Background(), zap.NewNop().Sugar())
		result, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), uuid.New())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))
	})

	It("constructs a correct request", func() {
		doer := withMockResponse(200, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), zap.NewNop().Sugar(), &doer)
		recipient := uuid.New()
		ctx := utils.SetLog(context.Background(), zap.NewNop().Sugar())
		result, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, uuid.New())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		bytes, err := ioutil.ReadAll(doer.request.Body)
		Expect(err).ToNot(HaveOccurred())
		parsedRequest := make(map[string]interface{})
		err = json.Unmarshal(bytes, &parsedRequest)
		Expect(err).ToNot(HaveOccurred())

		Expect(parsedRequest["account"]).To(Equal("1234"))
		Expect(parsedRequest["directive"]).To(Equal("playbook"))
		Expect(parsedRequest["recipient"]).To(Equal(recipient.String()))
	})

	// TODO: forwarded headers
})
