package utils

import (
	"fmt"
	"net/http"
)

func UnexpectedResponse(res *http.Response) error {
	return fmt.Errorf(`unexpected status code "%d" or content type "%s"`, res.StatusCode, res.Header.Get("content-type"))
}
