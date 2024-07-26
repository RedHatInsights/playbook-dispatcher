package utils

import (
	"fmt"
	"net/http"
)

type BlocklistedOrgIdError struct {
	OrgID string
}

func UnexpectedResponse(res *http.Response) error {
	return fmt.Errorf(`unexpected status code "%d" or content type "%s"`, res.StatusCode, res.Header.Get("content-type"))
}

func (this *BlocklistedOrgIdError) Error() string {
	return fmt.Sprintf("This org_id (%s) is blocklisted.", this.OrgID)
}
