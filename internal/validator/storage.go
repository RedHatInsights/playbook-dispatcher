package validator

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"playbook-dispatcher/internal/common/config"
	commonInstrumentation "playbook-dispatcher/internal/common/instrumentation"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/xi2/xz"
)

type storageConnector struct {
	client       utils.HttpRequestDoer
	retries      int
	timerFactory func() *prometheus.Timer
}

func newStorageConnector(cfg *viper.Viper) *storageConnector {
	return newStorageConnectorWithClient(cfg, &http.Client{
		Timeout: time.Duration(cfg.GetInt64("storage.timeout") * int64(time.Second)),
	})
}

func newStorageConnectorWithClient(cfg *viper.Viper, client utils.HttpRequestDoer) *storageConnector {
	return &storageConnector{
		client:       client,
		retries:      cfg.GetInt("storage.retries"),
		timerFactory: commonInstrumentation.OutboundHTTPDurationTimerFactory("storage"),
	}
}

func (this *storageConnector) initiateFetchWorkers(workers int, input <-chan messageContext, output chan<- enrichedMessageContext) {
	var workersWg sync.WaitGroup

	for i := 0; i < workers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()

			for {
				msg, open := <-input

				if !open {
					return
				}

				if payload, err := this.fetchPayload(msg.request.URL); err != nil {
					instrumentation.FetchArchiveError(msg.ctx, err, msg.requestType)
				} else {
					output <- enrichedMessageContext{messageContext: msg, data: payload}
				}
			}
		}()
	}

	workersWg.Wait()
	close(output)
}

func (this *storageConnector) fetchPayload(url string) (payload []byte, err error) {
	res, err := utils.DoGetWithRetry(this.client, url, this.retries, this.timerFactory)
	if err != nil {
		return
	}

	defer res.Body.Close()
	maxDecompressedSize := config.Get().GetInt64("artifact.max.decompressed.size")
	payload, err = readFile(res.Body, maxDecompressedSize)
	return
}

func readFile(reader io.Reader, maxDecompressedSize int64) (result []byte, err error) {
	// Validate limit is positive
	if maxDecompressedSize <= 0 {
		return nil, fmt.Errorf("invalid configuration: artifact.max.decompressed.size must be positive, got %d", maxDecompressedSize)
	}

	reader = bufio.NewReaderSize(reader, 2)
	compression, err := utils.GetCompressionType(reader)
	if err != nil {
		return nil, err
	}

	if compression == utils.GZip {
		if gzipReader, err := gzip.NewReader(reader); err != nil {
			return nil, err
		} else {
			defer gzipReader.Close()
			reader = gzipReader
		}
	}

	if compression == utils.XZ {
		if reader, err = xz.NewReader(reader, 0); err != nil {
			return nil, err
		}
	}

	// Read the file; if compressed, check against max size
	// Max artifact size was checked earlier so using maxDecompressedSize for all reads
	limitedReader := io.LimitReader(reader, maxDecompressedSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	// If we read up to exactly the limit, peek ahead one more byte
	if int64(len(data)) == maxDecompressedSize {
		// Try to read one more byte to detect oversized files
		// Assumes storage-backed readers return io.EOF when exhausted, not (0, nil) loops.
		var extra [1]byte
		n, err := reader.Read(extra[:])

		// Check for read errors first
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading decompressed archive: %w", err)
		}

		// Then check if file exceeds limit
		if n > 0 {
			return nil, fmt.Errorf("decompressed archive exceeds maximum allowed size (%d bytes)", maxDecompressedSize)
		}
	}

	return data, nil
}
