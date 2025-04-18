package db

import (
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/models"
	"gorm.io/gorm"
)

func DBMigrator(db *gorm.DB) error {
	return db.AutoMigrate(&models.Tool{}, &models.Category{}, &models.Step{})
}
