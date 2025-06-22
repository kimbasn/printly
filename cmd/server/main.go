package main

import (
	"log"

	"github.com/kimbasn/printly/internal/config"
	"github.com/kimbasn/printly/internal/db"
)

func main() {
	cfg := config.Load()
	db.Init(cfg)

	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Printf("🚀 Starting Printly in %s mode on port %s...\n", cfg.AppEnv, cfg.Port)
}
