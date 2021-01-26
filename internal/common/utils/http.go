package utils

import "net/http"

func DoGetWithRetry(client *http.Client, url string, retries int) (resp *http.Response, err error) {
	for ; retries > 0; retries-- {
		resp, err = client.Get(url)

		if err == nil {
			break
		}
	}

	return
}
