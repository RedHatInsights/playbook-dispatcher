package test

import (
	"fmt"
	"playbook-dispatcher/internal/common/config"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	. "github.com/onsi/gomega"
)

func WithAccountNumber() func() string {
	var base = uuid.New().String()[29:]
	var test int

	BeforeEach(func() {
		test++
	})

	return func() string {
		return fmt.Sprintf("%s-%d", base, test)
	}
}

func WithOrgId() func() string {
	var base = uuid.New().String()[29:]
	var test int

	BeforeEach(func() {
		test++
	})

	return func() string {
		return fmt.Sprintf("%s-%d", base, test)
	}
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
