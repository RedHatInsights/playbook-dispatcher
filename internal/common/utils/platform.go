package utils

import (
	"encoding/base64"
	"encoding/json"

	identityMiddleware "github.com/redhatinsights/platform-go-middlewares/identity"
)

func ParseIdentityHeader(raw string) (identity identityMiddleware.XRHID, err error) {
	identityRaw, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return
	}

	err = json.Unmarshal(identityRaw, &identity)
	return
}
