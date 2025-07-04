package db

import (
	"fmt"
	"log"

	"github.com/kimbasn/printly/internal/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBDriver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DBSource)
	// case "postgres":
	//	dialector = postgres.Open(cfg.DBSource)
	default:
		return nil, fmt.Errorf("Unsupported DB driver: %s", cfg.DBDriver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("Failed to get DB instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping DB: %w", err)
	}

	log.Println("✅ Connected to database:", cfg.DBDriver)
	return db, nil
}
