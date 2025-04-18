package db

import (
	"fmt"

	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/config"
	"github.com/gofiber/fiber/v2/log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDatabase(config *config.EnvConfig, DBMigrator func(*gorm.DB) error) *gorm.DB {
	uri := fmt.Sprintf(`host=%s user=%s password=%s dbname=%s port=%s sslmode=%s`,
		config.DBConfig.DBHost,
		config.DBConfig.DBUser,
		config.DBConfig.DBPassword,
		config.DBConfig.DBName,
		config.DBConfig.DBPort,
		config.DBConfig.DBSSLMode,
	)

	log.Info("Connecting to database with uri: ", uri)

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	log.Info("Connected to the database")

	if err := DBMigrator(db); err != nil {
		log.Fatalf("Unable to migrate: %v", err)
	}

	return db
}
