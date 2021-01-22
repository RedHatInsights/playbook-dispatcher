package test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

func IdentityHeaderMinimal(account string) string {
	data := fmt.Sprintf(`{"identity":{"internal":{"org_id":"12345"},"account_number":"%s","user":{},"type":"User"}}`, account)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

var Client = http.Client{
	Timeout: 1 * time.Second,
}
