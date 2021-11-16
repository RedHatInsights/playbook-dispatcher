package validator

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	commonInstrumentation "playbook-dispatcher/internal/common/instrumentation"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/validator/instrumentation"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
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
				select {
				case msg, open := <-input:
					if !open {
						return
					}

					if payload, err := this.fetchPayload(msg.request.URL); err != nil {
						instrumentation.FetchArchiveError(msg.ctx, err)
					} else {
						output <- enrichedMessageContext{messageContext: msg, data: payload}
					}
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
	payload, err = readFile(res.Body)
	return
}

func readFile(reader io.Reader) (result []byte, err error) {
	var isGzip bool
	reader = bufio.NewReaderSize(reader, 2)

	if isGzip, err = utils.IsGzip(reader); err != nil {
		return
	} else if isGzip {
		if gzipReader, err := gzip.NewReader(reader); err != nil {
			return nil, err
		} else {
			defer gzipReader.Close()
			reader = gzipReader
		}
	}

	return ioutil.ReadAll(reader)
}
