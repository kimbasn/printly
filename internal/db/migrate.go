package db

import (
	"github.com/kimbasn/printly/internal/entity"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Order{},
		&entity.PrintCenter{},
		&entity.Document{},
		&entity.Service{},
		&entity.WorkingHour{},
	)
}
