package validator

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"
	"strings"
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

	Describe("Read File", func() {
		It("Reads uncompressed file within limit", func() {
			content := "test data"
			reader := strings.NewReader(content)

			data, err := readFile(reader, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(content))
		})

		It("Rejects uncompressed file exceeding limit", func() {
			content := strings.Repeat("x", 200)
			reader := strings.NewReader(content)

			data, err := readFile(reader, 100)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exceeds maximum allowed size"))
			Expect(data).To(BeNil())
		})

		It("Reads gzip compressed file within limit", func() {
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			_, err := gzWriter.Write([]byte("test data"))
			Expect(err).ToNot(HaveOccurred())
			err = gzWriter.Close()
			Expect(err).ToNot(HaveOccurred())

			data, err := readFile(&buf, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("test data"))
		})

		It("Rejects gzip compressed file exceeding decompressed limit", func() {
			// Create large content that compresses well
			largeContent := strings.Repeat("a", 200)
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			_, err := gzWriter.Write([]byte(largeContent))
			Expect(err).ToNot(HaveOccurred())
			err = gzWriter.Close()
			Expect(err).ToNot(HaveOccurred())

			data, err := readFile(&buf, 100)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exceeds maximum allowed size"))
			Expect(data).To(BeNil())
		})

		It("Rejects invalid negative limit", func() {
			reader := strings.NewReader("test")

			data, err := readFile(reader, -1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must be positive"))
			Expect(data).To(BeNil())
		})

		It("Rejects zero limit", func() {
			reader := strings.NewReader("test")

			data, err := readFile(reader, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must be positive"))
			Expect(data).To(BeNil())
		})

		It("Accepts uncompressed file exactly at the limit", func() {
			content := strings.Repeat("x", 100)
			reader := strings.NewReader(content)

			data, err := readFile(reader, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(data)).To(Equal(100))
		})

		It("Accepts gzip compressed file with decompressed content exactly at the limit", func() {
			// Create exactly 100 bytes of content
			content := strings.Repeat("x", 100)
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			_, err := gzWriter.Write([]byte(content))
			Expect(err).ToNot(HaveOccurred())
			err = gzWriter.Close()
			Expect(err).ToNot(HaveOccurred())

			data, err := readFile(&buf, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(data)).To(Equal(100))
			Expect(string(data)).To(Equal(content))
		})

		It("Handles read errors during decompression", func() {
			// Create a corrupted gzip stream: valid header, then error mid-stream
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			_, err := gzWriter.Write([]byte("some data"))
			Expect(err).ToNot(HaveOccurred())
			err = gzWriter.Close()
			Expect(err).ToNot(HaveOccurred())

			// Take the gzip header but corrupt the compressed data
			gzipBytes := buf.Bytes()
			corruptedReader := &corruptedGzipReader{
				data:  gzipBytes[:10], // Valid gzip header
				index: 0,
			}

			data, err := readFile(corruptedReader, 100)
			Expect(err).To(HaveOccurred())
			Expect(data).To(BeNil())
		})

		It("Handles read errors during overflow probe", func() {
			// Create a reader that returns exactly 100 bytes, then a non-EOF error
			exactReader := &exactSizeThenErrorReader{
				data: []byte(strings.Repeat("x", 100)),
				err:  io.ErrUnexpectedEOF,
			}

			data, err := readFile(exactReader, 100)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error reading decompressed archive"))
			Expect(data).To(BeNil())
		})
	})
})

// Helper for testing corrupted gzip stream
type corruptedGzipReader struct {
	data  []byte
	index int
}

func (c *corruptedGzipReader) Read(p []byte) (n int, err error) {
	if c.index >= len(c.data) {
		// After header, return error to simulate corrupted stream
		return 0, io.ErrUnexpectedEOF
	}
	n = copy(p, c.data[c.index:])
	c.index += n
	return n, nil
}

// Helper for testing error during overflow probe
type exactSizeThenErrorReader struct {
	data     []byte
	index    int
	err      error
	returned bool
}

func (e *exactSizeThenErrorReader) Read(p []byte) (n int, err error) {
	if e.index >= len(e.data) {
		// Data exhausted, now return error on overflow probe
		if !e.returned {
			e.returned = true
			return 0, e.err
		}
		return 0, io.EOF
	}
	n = copy(p, e.data[e.index:])
	e.index += n
	return n, nil
}
