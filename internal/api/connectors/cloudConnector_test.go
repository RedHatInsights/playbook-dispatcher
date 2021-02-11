package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/constants"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"
	"strconv"

	"github.com/redhatinsights/platform-go-middlewares/request_id"

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

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		result, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), uuid.New(), "http://example.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))
	})

	It("interprets the response correctly on bad request", func() {
		doer := withMockResponse(400, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		_, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), uuid.New(), "http://example.com")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`unexpected status code "400"`))
	})

	It("constructs a correct request", func() {
		doer := withMockResponse(200, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		url := "http://example.com"
		correlationId := uuid.New()

		returnUrl := "http://example.com/return"
		responseInterval := 60
		cfg := config.Get()
		cfg.Set("return.url", returnUrl)
		cfg.Set("response.interval", responseInterval)

		client := NewConnectorClientWithHttpRequestDoer(cfg, &doer)
		recipient := uuid.New()
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		result, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, correlationId, url)
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
		Expect(parsedRequest["payload"]).To(Equal(url))

		metadata, ok := parsedRequest["metadata"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(metadata["crc_dispatcher_correlation_id"]).To(Equal(correlationId.String()))
		Expect(metadata["return_url"]).To(Equal(returnUrl))
		Expect(metadata["response_interval"]).To(Equal(strconv.Itoa(responseInterval)))
	})

	It("forwards identity header", func() {
		requestId := "e6b06142-9589-4213-9a5e-1e2f513c448b"
		doer := withMockResponse(200, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)
		ctx := context.WithValue(test.TestContext(), request_id.RequestIDKey, requestId)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		recipient := uuid.New()
		result, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, uuid.New(), "http://example.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		idHeader := doer.request.Header.Get(constants.HeaderRequestId)
		Expect(idHeader).To(Equal(requestId))
	})
})
