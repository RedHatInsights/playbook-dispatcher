package tests

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils/test"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg := config.Get()
	cfg.Set("web.port", 9002)

	api.WithApi(cfg)

	RunSpecs(t, "Controllers Suite")
}

var (
	accountNumber = test.WithAccountNumber()

	client = &Client{
		Server: "http://localhost:9002",
		Client: &test.Client,
		RequestEditor: func(ctx context.Context, req *http.Request) error {
			if account := ctx.Value(accountContextKey); account != nil {
				req.Header.Set("x-rh-identity", test.IdentityHeaderMinimal(account.(string)))
			}

			return nil
		},
	}
)

const accountContextKey = iota

func ContextWithIdentity(account string) context.Context {
	return context.WithValue(context.Background(), accountContextKey, account)
}

func WithDatabase() func() *gorm.DB {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		cfg := config.Get()
		dsn := fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.GetString("db.username"),
			cfg.GetString("db.password"),
			cfg.GetString("db.host"),
			cfg.GetInt("db.port"),
			cfg.GetString("db.name"),
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if sqlConnection, err := db.DB(); err != nil {
			sqlConnection.Close()
		}
	})

	return func() *gorm.DB {
		return db
	}
}
