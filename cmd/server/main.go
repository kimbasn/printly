package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/kimbasn/printly/docs" // Swagger docs
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/kimbasn/printly/internal/config"
	"github.com/kimbasn/printly/internal/db"
	"github.com/kimbasn/printly/internal/middlewares"
	"github.com/kimbasn/printly/internal/routes"
	"github.com/kimbasn/printly/internal/service"

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
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load and validate configuration
	cfg := config.Load()
	cfg.PrintConfig()

	// Validate configuration
	if err := cfg.ValidateConfig(); err != nil {
		logger.Fatal("Configuration validation failed", zap.Error(err))
	}

	// Initialize database
	dbConn, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Database initialization failed", zap.Error(err))
	}

	// Initialize Firebase
	firebaseApp, err := initFirebase(cfg, logger)
	if err != nil {
		logger.Fatal("Firebase initialization failed", zap.Error(err))
	}

	// Initialize storage service
	storageService, err := service.GetStorageService(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize storage service", zap.Error(err))
	}

	// Setup server
	server := setupServer(cfg, dbConn, firebaseApp, storageService, logger)

	// Start server with graceful shutdown
	startServerWithGracefulShutdown(server, cfg, logger)
}

func initLogger() (*zap.Logger, error) {
	// You might want to configure different loggers for different environments
	return zap.NewProduction()
}

func initDatabase(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	logger.Info("Initializing database connection...")

	dbConn, err := db.Init(cfg)
	if err != nil {
		return nil, err
	}

	logger.Info("Running database migrations...")
	if err := db.AutoMigrate(dbConn); err != nil {
		return nil, err
	}

	logger.Info("Database initialized successfully")
	return dbConn, nil
}

func initFirebase(cfg *config.Config, logger *zap.Logger) (*firebase.App, error) {
	logger.Info("Initializing Firebase...")

	firebaseApp, err := config.SetupFirebase(context.Background(), cfg.FirebaseCredentialsFile)
	if err != nil {
		return nil, err
	}

	logger.Info("Firebase initialized successfully")
	return firebaseApp, nil
}

func setupServer(cfg *config.Config,
	dbConn *gorm.DB,
	firebaseApp *firebase.App,
	storageService service.StorageService,
	logger *zap.Logger) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := gin.New()

	// Add middleware
	server.Use(gin.Recovery())
	server.Use(middlewares.Logger())
	server.Use(cors.Default())

	// Health check endpoint
	server.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Swagger route (only in non-production environments)
	if cfg.AppEnv != "production" {
		server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API V1 grouping
	api := server.Group("/api/v1")
	validate := validator.New()

	// Register routes
	routes.RegisterUserRoutes(api, dbConn, validate, firebaseApp)
	routes.RegisterPrintCenterRoutes(api, dbConn, validate, firebaseApp)
	routes.RegisterOrderRoutes(api, dbConn, validate, firebaseApp, logger, storageService)

	logger.Info("Server setup completed")
	return server
}

func startServerWithGracefulShutdown(server *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	serverAddress := cfg.GetServerAddress()

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: server,
		// Add timeouts for security
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Printly server",
			zap.String("environment", cfg.AppEnv),
			zap.String("address", serverAddress))

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
