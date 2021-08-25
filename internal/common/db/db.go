package db

import (
	"context"
	"database/sql"
	"fmt"
	"playbook-dispatcher/internal/common/utils"
	"time"

	"go.uber.org/zap"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(ctx context.Context, cfg *viper.Viper) (*gorm.DB, *sql.DB) {
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

	log := utils.GetLogFromContext(ctx)
	log.Infow("Connecting to database", "host", cfg.GetString("db.host"), "sslmode", cfg.GetString("db.sslmode"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: &zapAdapter{
			log: log.Named("gorm"),
		},
	})

	utils.DieOnError(err)

	sql, err := db.DB()
	utils.DieOnError(err)

	sql.SetMaxIdleConns(cfg.GetInt("db.max.idle.connections"))
	sql.SetMaxOpenConns(cfg.GetInt("db.max.open.connections"))

	return db, sql
}

type zapAdapter struct {
	log        *zap.SugaredLogger
	currentLog *zap.SugaredLogger
}

func (this *zapAdapter) getLog() *zap.SugaredLogger {
	if this.currentLog != nil {
		return this.currentLog
	}

	return this.log
}

func (this *zapAdapter) Info(ctx context.Context, msg string, values ...interface{}) {
	this.getLog().Infow(msg, values...)
}

func (this *zapAdapter) Warn(ctx context.Context, msg string, values ...interface{}) {
	this.getLog().Warnw(msg, values...)
}

func (this *zapAdapter) Error(ctx context.Context, msg string, values ...interface{}) {
	this.getLog().Errorw(msg, values...)
}

func (this *zapAdapter) LogMode(level logger.LogLevel) logger.Interface {
	return this
}

func (this *zapAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	this.getLog().Debugw("executed query", "sql", sql, "rows", rows, "elapsed", elapsed.Milliseconds())
}

func SetLog(db *gorm.DB, log *zap.SugaredLogger) {
	if adapter, ok := db.Logger.(*zapAdapter); ok {
		adapter.currentLog = log
	}
}

func ClearLog(db *gorm.DB) {
	SetLog(db, nil)
}
