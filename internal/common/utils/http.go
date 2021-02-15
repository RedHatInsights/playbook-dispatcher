package utils

import (
	"bufio"
	"io"
	"net/http"
)

const (
	// https://tools.ietf.org/html/rfc1952#section-2.3.1
	gzipByte1 = 0x1f
	gzipByte2 = 0x8b
)

func DoGetWithRetry(client *http.Client, url string, retries int) (resp *http.Response, err error) {
	for ; retries > 0; retries-- {
		resp, err = client.Get(url)

		if err == nil {
			break
		}
	}

	return
}

func IsGzip(reader io.Reader) (bool, error) {
	bufferedReader := bufio.NewReaderSize(reader, 2)
	peek, err := bufferedReader.Peek(2)

	if err != nil {
		return false, err
	}

	return peek[0] == gzipByte1 && peek[1] == gzipByte2, nil
}
