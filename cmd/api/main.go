package main

import (
	"fmt"

	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/config"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/db"
	_ "github.com/can4hou6joeng4/convenient-tools-project-v1-backend/docs" // 导入swagger文档
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/handlers"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/repositories"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Convenient Tools API
// @version 1.0
// @description This is a convenient tools project API documentation.

// @contact.name bobochang
// @contact.url https://github.com/can4hou6joeng4/
// @contact.email can4hou6joeng4@163.com

// @host localhost:8082
// @BasePath /api

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "ConvenientTools",
		ServerHeader: "Fiber",
	})

	// Config
	envConfig := config.NewEnvConfig()
	redis := db.InitRedis(envConfig)
	cos := db.InitCOSClient(envConfig)
	db := db.InitDatabase(envConfig, db.DBMigrator)

	// Swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Repository
	toolRepository := repositories.NewToolRepository(db)

	// Routing
	server := app.Group("/api")
	handlers.NewCommonHandler(server, toolRepository, redis, cos, envConfig)

	app.Listen(fmt.Sprintf(":%s", envConfig.ServerPort))

}
