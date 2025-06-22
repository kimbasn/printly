package db

import (
	"log"

	"github.com/kimbasn/printly/internal/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	var dialector gorm.Dialector

	switch cfg.DBDriver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DBSource)
	// case "postgres":
	//	dialector = postgres.Open(cfg.DBSource)
	default:
		log.Fatalf("Unsupported DB driver: %s", cfg.DBDriver)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get DB instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping DB: %v", err)
	}

	log.Println("✅ Connected to database:", cfg.DBDriver)
}
