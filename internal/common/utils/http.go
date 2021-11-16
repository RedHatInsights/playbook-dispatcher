package utils

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpCallback func(req *http.Request) (status int, body string, err error)

type mockHttpRequestDoer struct {
	callback httpCallback
}

const (
	// https://tools.ietf.org/html/rfc1952#section-2.3.1
	gzipByte1 = 0x1f
	gzipByte2 = 0x8b
)

func DoGetWithRetry(client HttpRequestDoer, url string, retries int, timerFactory func() *prometheus.Timer) (resp *http.Response, err error) {
	for ; retries > 0; retries-- {
		resp, err = doGet(client, url, timerFactory)

		if err == nil {
			break
		}
	}

	return
}

func doGet(client HttpRequestDoer, url string, timerFactory func() *prometheus.Timer) (resp *http.Response, err error) {
	if timerFactory != nil {
		timer := timerFactory()
		defer timer.ObserveDuration()
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func IsGzip(reader io.Reader) (bool, error) {
	bufferedReader := bufio.NewReaderSize(reader, 2)
	peek, err := bufferedReader.Peek(2)

	if err != nil {
		return false, err
	}

	return peek[0] == gzipByte1 && peek[1] == gzipByte2, nil
}

func (this *mockHttpRequestDoer) Do(req *http.Request) (*http.Response, error) {
	status, body, error := this.callback(req)
	return &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}, error
}

func NewMockHttpRequestDoer(status int, body string, err error) HttpRequestDoer {
	s, b, e := status, body, err
	return &mockHttpRequestDoer{
		callback: func(req *http.Request) (status int, body string, err error) {
			return s, b, e
		},
	}
}

func NewMockHttpRequestDoerWithCallback(callback httpCallback) HttpRequestDoer {
	return &mockHttpRequestDoer{
		callback: callback,
	}
}
