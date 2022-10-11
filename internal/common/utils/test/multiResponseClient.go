package test

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type MockHttpResponse struct {
	StatusCode int
	Body       string
}

type mockMultiResponseHttpRequestDoer struct {
	Request      *http.Request
	responses    []http.Response
	requestCount int
}

func (this *mockMultiResponseHttpRequestDoer) Do(req *http.Request) (*http.Response, error) {
	this.Request = req
	response := this.responses[this.requestCount]
	this.requestCount++
	return &response, nil
}

func MockMultiResponseHttpClient(mockResponses ...MockHttpResponse) *mockMultiResponseHttpRequestDoer {

	responseList := make([]http.Response, 0, len(mockResponses))

	for i := range mockResponses {
		response := http.Response{
			StatusCode: mockResponses[i].StatusCode,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(mockResponses[i].Body))),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}

		responseList = append(responseList, response)
	}

	return &mockMultiResponseHttpRequestDoer{
		responses: responseList,
	}
}
