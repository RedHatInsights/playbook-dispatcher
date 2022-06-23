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

var (
	ansibleDirective = "playbook"
	satDirective     = "playbook-sat"
)

type CustomUrlType struct {
	Payload json.RawMessage
}

func ansibleMetadata(correlationId uuid.UUID) map[string]string {
	return map[string]string{
		"crc_dispatcher_correlation_id": correlationId.String(),
		"return_url":                    "http://example.com/return",
		"response_interval":             "60",
	}
}

var _ = Describe("Cloud Connector", func() {
	It("interprets the response correctly", func() {
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		correlationId := uuid.New()
		url := "http://example.com"
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), &url, ansibleDirective, ansibleMetadata(correlationId))
		Expect(notFound).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))
	})

	It("interprets the response correctly on bad request", func() {
		doer := test.MockHttpClient(400, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		correlationId := uuid.New()
		url := "http://example.com"
		_, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), &url, ansibleDirective, ansibleMetadata(correlationId))
		Expect(notFound).To(BeFalse())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`unexpected status code "400"`))
	})

	It("interprets the response correctly on no connection found", func() {
		doer := test.MockHttpClient(404, `{}`)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		correlationId := uuid.New()
		url := "http://example.com"
		id, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", uuid.New(), &url, ansibleDirective, ansibleMetadata(correlationId))
		Expect(id).To(BeNil())
		Expect(notFound).To(BeTrue())
		Expect(err).ToNot(HaveOccurred())
	})

	It("constructs a correct request", func() {
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		url := "http://example.com"
		correlationId := uuid.New()

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		recipient := uuid.New()
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, &url, ansibleDirective, ansibleMetadata(correlationId))
		Expect(notFound).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		bytes, err := ioutil.ReadAll(doer.Request.Body)
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
		Expect(metadata["return_url"]).To(Equal("http://example.com/return"))
		Expect(metadata["response_interval"]).To(Equal(strconv.Itoa(60)))
	})

	It("constructs a correct satellite request", func() {
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		url := "http://example.com"
		correlationId := uuid.New()

		satMetadata := map[string]string{
			"operation":         "run",
			"correlation_id":    correlationId.String(),
			"playbook_run_name": "test-playbook",
			"playbook_run_url":  "http://example.com",
			"sat_id":            "16372e6f-1c18-4cdb-b780-50ab4b88e74b",
			"sat_org_id":        "123",
			"initiator_user_id": "test-user",
			"hosts":             "16372e6f-1c18-4cdb-b780-50ab4b88e74b,baf2bb2f-06a3-42cc-ae7b-68ccc8e2a344",
			"return_url":        "http://example.com/return",
			"response_interval": "60",
		}

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		recipient := uuid.New()
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, &url, satDirective, satMetadata)
		Expect(notFound).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		bytes, err := ioutil.ReadAll(doer.Request.Body)
		Expect(err).ToNot(HaveOccurred())
		parsedRequest := make(map[string]interface{})
		err = json.Unmarshal(bytes, &parsedRequest)
		Expect(err).ToNot(HaveOccurred())

		Expect(parsedRequest["account"]).To(Equal("1234"))
		Expect(parsedRequest["directive"]).To(Equal("playbook-sat"))
		Expect(parsedRequest["recipient"]).To(Equal(recipient.String()))
		Expect(parsedRequest["payload"]).To(Equal(url))

		metadata, ok := parsedRequest["metadata"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(metadata["return_url"]).To(Equal("http://example.com/return"))
		Expect(metadata["response_interval"]).To(Equal(strconv.Itoa(60)))

		Expect(metadata["operation"]).To(Equal("run"))
		Expect(metadata["correlation_id"]).To(Equal(correlationId.String()))
		Expect(metadata["playbook_run_name"]).To(Equal(satMetadata["playbook_run_name"]))
		Expect(metadata["playbook_run_url"]).To(Equal(satMetadata["playbook_run_url"]))
		Expect(metadata["sat_id"]).To(Equal(satMetadata["sat_id"]))
		Expect(metadata["sat_org_id"]).To(Equal(satMetadata["sat_org_id"]))
		Expect(metadata["initiator_user_id"]).To(Equal(satMetadata["initiator_user_id"]))
		Expect(metadata["hosts"]).To(Equal(satMetadata["hosts"]))
	})

	It("constructs a correct satellite cancel request", func() {
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		correlationId := uuid.New()

		satCancelMetadata := map[string]string{
			"operation":         "cancel",
			"correlation_id":    correlationId.String(),
			"initiator_user_id": "test-user",
		}

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		recipient := uuid.New()
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, nil, satDirective, satCancelMetadata)
		Expect(notFound).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		bytes, err := ioutil.ReadAll(doer.Request.Body)
		Expect(err).ToNot(HaveOccurred())
		parsedRequest := make(map[string]interface{})
		err = json.Unmarshal(bytes, &parsedRequest)
		Expect(err).ToNot(HaveOccurred())

		Expect(parsedRequest).To(HaveLen(4))
		Expect(parsedRequest["account"]).To(Equal("1234"))
		Expect(parsedRequest["directive"]).To(Equal("playbook-sat"))
		Expect(parsedRequest["recipient"]).To(Equal(recipient.String()))

		metadata, ok := parsedRequest["metadata"].(map[string]interface{})
		Expect(ok).To(BeTrue())

		Expect(metadata["operation"]).To(Equal("cancel"))
		Expect(metadata["correlation_id"]).To(Equal(correlationId.String()))
		Expect(metadata["initiator_user_id"]).To(Equal(satCancelMetadata["initiator_user_id"]))
	})

	It("forwards identity header", func() {
		requestId := "e6b06142-9589-4213-9a5e-1e2f513c448b"
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)
		ctx := context.WithValue(test.TestContext(), request_id.RequestIDKey, requestId)

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		recipient := uuid.New()
		correlationId := uuid.New()
		url := "http://example.com"
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, &url, ansibleDirective, ansibleMetadata(correlationId))
		Expect(err).ToNot(HaveOccurred())
		Expect(notFound).To(BeFalse())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		idHeader := doer.Request.Header.Get(constants.HeaderRequestId)
		Expect(idHeader).To(Equal(requestId))
	})

	It("does not escape ampersands with unicode", func() {
		doer := test.MockHttpClient(201, `{"id": "871e31aa-7d41-43e3-8ef7-05706a0ee34a"}`)

		url := "http://example.com/?field1=test&field2=test2&field3"
		correlationId := uuid.New()

		client := NewConnectorClientWithHttpRequestDoer(config.Get(), &doer)
		recipient := uuid.New()
		ctx := utils.SetLog(test.TestContext(), zap.NewNop().Sugar())
		result, notFound, err := client.SendCloudConnectorRequest(ctx, "1234", recipient, &url, ansibleDirective, ansibleMetadata(correlationId))
		Expect(notFound).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
		Expect(*result).To(Equal("871e31aa-7d41-43e3-8ef7-05706a0ee34a"))

		bytes, err := ioutil.ReadAll(doer.Request.Body)
		Expect(err).ToNot(HaveOccurred())
		parsedRequest := &CustomUrlType{}
		err = json.Unmarshal(bytes, parsedRequest)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(parsedRequest.Payload)).To(Equal("\"http://example.com/?field1=test&field2=test2&field3\""))
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
