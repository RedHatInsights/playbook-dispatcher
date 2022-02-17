package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/constants"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"
	"strconv"

	"github.com/redhatinsights/platform-go-middlewares/request_id"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Cloud Connector", func() {
	It("interprets the response correctly", func() {
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), uuid.New(), "http://example.com")
		Expect(notFound).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))
	})

	It("interprets the response correctly on bad request", func() {
		doer := test.MockHttpClient(400, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		_, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), uuid.New(), "http://example.com")
		Expect(notFound).To(BeFalse())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`unexpected status code "400"`))
	})

	It("interprets the response correctly on no connection found", func() {
		doer := test.MockHttpClient(404, `{}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		id, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), uuid.New(), "http://example.com")
		Expect(id).To(BeNil())
		Expect(notFound).To(BeTrue())
		Expect(err).ToNot(HaveOccurred())
	})

	It("constructs a correct request", func() {
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

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
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, correlationId, url)
		Expect(notFound).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		bytes, err := ioutil.ReadAll(doer.Request.Body)
		Expect(err).ToNot(HaveOccurred())
		parsedRequest := make(map[string]interface{})
		err = json.Unmarshal(bytes, &parsedRequest)
		Expect(err).ToNot(HaveOccurred())

		Expect(parsedRequest["account"]).To(Equal("1234"))
		Expect(parsedRequest["directive"]).To(Equal("rhc-worker-playbook"))
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
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)
		ctx := context.WithValue(test.TestContext(), request_id.RequestIDKey, requestId)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		recipient := uuid.New()
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, uuid.New(), "http://example.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(notFound).To(BeFalse())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		idHeader := doer.Request.Header.Get(constants.HeaderRequestId)
		Expect(idHeader).To(Equal(requestId))
	})

	Describe("connection status", func() {
		DescribeTable("interprets the response correctly",
			func(status string, expectedStatus ConnectionStatus) {
				doer := test.MockHttpClient(200, fmt.Sprintf(`{"status": "%s"}`, status))

				client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
				ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())

				result, err := client.GetConnectionStatus(ctx, "901578", "5318290", "be175f04-4634-49f2-a292-b4ad7107af78")
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(expectedStatus))
			},

			Entry("connected", "connected", ConnectionStatus_connected),
			Entry("disconnected", "disconnected", ConnectionStatus_disconnected),
		)

		It("constructs a correct request", func() {
			doer := test.MockHttpClient(200, `{"status": "connected"}`)

			client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
			ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
			_, err := client.GetConnectionStatus(ctx, "901578", "5318290", "be175f04-4634-49f2-a292-b4ad7107af78")
			Expect(err).ToNot(HaveOccurred())

			bytes, err := ioutil.ReadAll(doer.Request.Body)
			Expect(err).ToNot(HaveOccurred())
			parsedRequest := make(map[string]interface{})
			err = json.Unmarshal(bytes, &parsedRequest)
			Expect(err).ToNot(HaveOccurred())

			Expect(parsedRequest["account"]).To(Equal("901578"))
			Expect(parsedRequest["node_id"]).To(Equal("be175f04-4634-49f2-a292-b4ad7107af78"))

			Expect(doer.Request.Header.Get(constants.HeaderCloudConnectorAccount)).To(Equal("901578"))
			Expect(doer.Request.Header.Get(constants.HeaderCloudConnectorOrgID)).To(Equal("5318290"))
		})
	})
})
