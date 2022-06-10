package test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/utils"
	"time"

	"go.uber.org/zap"
)

func IdentityHeaderMinimal(account string) string {
	data := fmt.Sprintf(`{"identity":{"internal":{"org_id":"%s"},"account_number":"%s","user":{},"type":"User"}}`, account+"-test", account)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

var Client = http.Client{
	Timeout: 1 * time.Second,
}

func TestContext() context.Context {
	return utils.SetLog(context.Background(), zap.NewNop().Sugar())
}
