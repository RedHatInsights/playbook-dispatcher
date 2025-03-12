package test

import (
	"bytes"
	"io"
	"net/http"
)

type mockHttpRequestDoer struct {
	Request  *http.Request
	response *http.Response
}

func (this *mockHttpRequestDoer) Do(req *http.Request) (*http.Response, error) {
	this.Request = req
	return this.response, nil
}

func MockHttpClient(statusCode int, body string) mockHttpRequestDoer {
	response := http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	return mockHttpRequestDoer{
		response: &response,
	}
}
