package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/kimbasn/printly/docs" // Swagger docs

	"github.com/kimbasn/printly/internal/config"
	"github.com/kimbasn/printly/internal/db"
	"github.com/kimbasn/printly/internal/routes"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/go-playground/validator/v10"
)

// @title           Printly API
// @version         1.0
// @description     API documentation for the Printly document printing platform.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Kimba SABI N'GOYE
// @contact.email  kimbasabingoye@printly.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	cfg := config.Load()
	dbConn, err := db.Init(cfg)
	if err != nil {
		log.Fatalf("❌ Could not initialize database: %v", err)
	}

	if err := db.AutoMigrate(dbConn); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Printf("🚀 Starting Printly in %s mode on port %s...\n", cfg.AppEnv, cfg.Port)

	router := gin.Default()

	// Allow CORS for frontend apps
	router.Use(cors.Default())

	// Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))


	// API V1 grouping
	api := router.Group("api/v1")
	routes.RegisterUserRoutes(api, dbConn, validator.New())

	// Start server
	serverAddress := cfg.Host + ":" + cfg.Port
	if err := router.Run(serverAddress); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}

}
