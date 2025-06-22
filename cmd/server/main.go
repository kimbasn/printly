package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kimbasn/printly/internal/config"
	"github.com/kimbasn/printly/internal/db"
	"github.com/kimbasn/printly/internal/routes"
)

func main() {
	cfg := config.Load()
	db.Init(cfg)

	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Printf("🚀 Starting Printly in %s mode on port %s...\n", cfg.AppEnv, cfg.Port)

	router := gin.Default()

	api := router.Group("api/v1")
	routes.RegisterUserRoutes(api)

	serverAddress := cfg.Host + ":" + cfg.Port
	router.Run(serverAddress)

}
