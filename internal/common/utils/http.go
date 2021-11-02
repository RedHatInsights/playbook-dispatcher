package utils

import (
	"bufio"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// https://tools.ietf.org/html/rfc1952#section-2.3.1
	gzipByte1 = 0x1f
	gzipByte2 = 0x8b
)

func DoGetWithRetry(client *http.Client, url string, retries int, timerFactory func() *prometheus.Timer) (resp *http.Response, err error) {
	for ; retries > 0; retries-- {
		resp, err = doGet(client, url, timerFactory)

		if err == nil {
			break
		}
	}

	return
}

func doGet(client *http.Client, url string, timerFactory func() *prometheus.Timer) (resp *http.Response, err error) {
	if timerFactory != nil {
		timer := timerFactory()
		defer timer.ObserveDuration()
	}

	return client.Get(url)
}

func IsGzip(reader io.Reader) (bool, error) {
	bufferedReader := bufio.NewReaderSize(reader, 2)
	peek, err := bufferedReader.Peek(2)

	if err != nil {
		return false, err
	}

	return peek[0] == gzipByte1 && peek[1] == gzipByte2, nil
}
