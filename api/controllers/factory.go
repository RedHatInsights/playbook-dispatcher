package controllers

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateControllers(database *gorm.DB, log *zap.SugaredLogger) ServerInterface {
	return &controllers{
		database: database,
		log:      log,
	}
}

// implements api.ServerInterface
type controllers struct {
	database *gorm.DB
	log      *zap.SugaredLogger
}
