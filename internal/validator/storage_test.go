package validator

import (
	"context"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Storage", func() {
	Describe("Fetch Payload", func() {
		It("Issues an HTTP GET call", func() {
			client := utils.NewMockHttpRequestDoer(200, "test", nil)
			storage := newStorageConnectorWithClient(config.Get(), client)

			response, err := storage.fetchPayload("http://example.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(response)).To(Equal("test"))
		})
	})

	Describe("Workers", func() {
		It("Fetches payloads concurrently", func() {
			concurrency := 10

			inputChan, outputChan := make(chan messageContext), make(chan enrichedMessageContext)
			client := utils.NewMockHttpRequestDoerWithCallback(func(req *http.Request) (status int, body string, err error) {
				time.Sleep(200 * time.Millisecond)
				return 200, "test", nil
			})
			storage := newStorageConnectorWithClient(config.Get(), client)

			go storage.initiateFetchWorkers(concurrency, inputChan, outputChan)

			start := time.Now()

			for i := 0; i < concurrency; i++ {
				inputChan <- messageContext{
					request: message.IngressValidationRequest{},
					ctx:     utils.SetLog(context.Background(), zap.NewNop().Sugar()),
				}
			}

			close(inputChan)

			for i := 0; i < concurrency; i++ {
				result, ok := <-outputChan
				Expect(ok).To(BeTrue())
				Expect(string(result.data)).To(Equal("test"))
			}

			_, ok := <-outputChan
			Expect(ok).To(BeFalse())

			end := time.Since(start)
			Expect(end).To(BeNumerically("<", time.Second))
		})
	})
})
