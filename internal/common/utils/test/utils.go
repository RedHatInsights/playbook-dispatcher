package test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/utils"
	"time"

	dbModel "playbook-dispatcher/internal/common/model/db"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func IdentityHeaderMinimal(account string) string {
	data := fmt.Sprintf(`{"identity":{"internal":{"org_id":"12345"},"account_number":"%s","user":{},"type":"User"}}`, account)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

var Client = http.Client{
	Timeout: 1 * time.Second,
}

func NewRunWithStatus(account string, status string) *dbModel.Run {
	return &dbModel.Run{
		ID:            uuid.New(),
		Account:       account,
		Recipient:     uuid.New(),
		CorrelationID: uuid.New(),
		URL:           "http://example.com",
		Status:        status,
		Timeout:       3600,
	}
}

func NewRun(account string) *dbModel.Run {
	return NewRunWithStatus(account, "running")
}

func TestContext() context.Context {
	return utils.SetLog(context.Background(), zap.NewNop().Sugar())
}
