package db

import (
	"github.com/kimbasn/printly/internal/entity"
)

func AutoMigrate() error {
	return DB.AutoMigrate(
		&entity.User{},
		&entity.Order{},
		&entity.PrintCenter{},
		&entity.Document{},
		&entity.Location{},
		&entity.GeoPoint{},
		&entity.Service{},
		&entity.WorkingHour{},
	)
}