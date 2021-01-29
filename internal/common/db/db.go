package db

import (
	"database/sql"
	"fmt"
	"playbook-dispatcher/internal/common/utils"

	"go.uber.org/zap"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *viper.Viper, log *zap.SugaredLogger) (*gorm.DB, *sql.DB) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.GetString("db.host"),
		cfg.GetInt("db.port"),
		cfg.GetString("db.name"),
		cfg.GetString("db.username"),
		cfg.GetString("db.password"),
		cfg.GetString("db.sslmode"),
	)

	if cfg.IsSet("db.ca") {
		dsn += fmt.Sprintf(" sslrootcert=%s", cfg.GetString("db.ca"))
	}

	log.Infow("Connecting to database", "host", cfg.GetString("db.host"), "sslmode", cfg.GetString("db.sslmode"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	utils.DieOnError(err)

	sql, err := db.DB()
	utils.DieOnError(err)

	sql.SetMaxIdleConns(cfg.GetInt("db.max.idle.connections"))
	sql.SetMaxOpenConns(cfg.GetInt("db.max.open.connections"))

	return db, sql
}
