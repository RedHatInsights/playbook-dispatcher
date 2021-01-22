package db

import (
	"database/sql"
	"fmt"
	"playbook-dispatcher/internal/common/utils"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *viper.Viper) (*gorm.DB, *sql.DB) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.GetString("db.username"),
		cfg.GetString("db.password"),
		cfg.GetString("db.host"),
		cfg.GetInt("db.port"),
		cfg.GetString("db.name"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})

	utils.DieOnError(err)

	sql, err := db.DB()
	utils.DieOnError(err)

	sql.SetMaxIdleConns(cfg.GetInt("db.max.idle.connections"))
	sql.SetMaxOpenConns(cfg.GetInt("db.max.open.connections"))

	return db, sql
}
