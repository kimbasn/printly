package main

import (
	"context"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/kimbasn/printly/docs" // Swagger docs

	"github.com/kimbasn/printly/internal/config"
	"github.com/kimbasn/printly/internal/db"
	"github.com/kimbasn/printly/internal/middlewares"
	"github.com/kimbasn/printly/internal/routes"

	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a JWT token.
func main() {
	cfg := config.Load()
	dbConn, err := db.Init(cfg)
	if err != nil {
		log.Fatalf("❌ Could not initialize database: %v", err)
	}

	if err := db.AutoMigrate(dbConn); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	firebaseApp, err := config.SetupFirebase(context.Background(), cfg.FirebaseCredentialsFile)
	if err != nil {
		log.Fatalf("❌ Could not initialize Firebase: %v", err)
	}

	log.Printf("🚀 Starting Printly in %s mode on port %s...\n", cfg.AppEnv, cfg.Port)

	// server := gin.Default()
	server := gin.New()

	server.Use(gin.Recovery())

	server.Use(middlewares.Logger())

	// Allow CORS for frontend apps
	server.Use(cors.Default())

	// Swagger route
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))


	// API V1 grouping
	api := server.Group("api/v1")
	validate := validator.New()
	routes.RegisterUserRoutes(api, dbConn, validate, firebaseApp)
	routes.RegisterPrintCenterRoutes(api, dbConn, validate, firebaseApp)
	routes.RegisterOrderRoutes(api, dbConn, validate, firebaseApp)

	// Start server
	serverAddress := cfg.Host + ":" + cfg.Port
	if err := server.Run(serverAddress); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}

}
