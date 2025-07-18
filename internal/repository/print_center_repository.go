package repository

import (
	"errors"
	"fmt"

	"github.com/kimbasn/printly/internal/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../mocks/mock_print_center_repository.go -package=mocks github.com/kimbasn/printly/internal/repository PrintCenterRepository

type PrintCenterRepository interface {
	Save(printCenter *entity.PrintCenter) error
	FindByID(id uint) (*entity.PrintCenter, error)
	FindByStatus(status entity.PrintCenterStatus) ([]entity.PrintCenter, error)
	FindAll() ([]entity.PrintCenter, error)
	Update(id uint, updates map[string]any) error
	Delete(id uint) error
}

type printCenterRepository struct {
	db *gorm.DB
}

func NewPrintCenterRepository(db *gorm.DB) PrintCenterRepository {
	return &printCenterRepository{db: db}
}

func (r *printCenterRepository) Save(printCenter *entity.PrintCenter) error {
	if err := r.db.Create(printCenter).Error; err != nil {
		return fmt.Errorf("failed to save the printing center: %w", err)
	}
	return nil
}

func (r *printCenterRepository) FindByID(id uint) (*entity.PrintCenter, error) {
	var printCenter entity.PrintCenter
	result := r.db.First(&printCenter, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to fetch print center with id %d: %w", id, result.Error)
	}
	return &printCenter, nil
}

func (r *printCenterRepository) FindByStatus(status entity.PrintCenterStatus) ([]entity.PrintCenter, error) {
	var printCenters []entity.PrintCenter
	result := r.db.Find(&printCenters, "status = ?", status)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch print centers with status: %s: %w", status, result.Error)
	}
	return printCenters, nil
}

func (r *printCenterRepository) FindAll() ([]entity.PrintCenter, error) {
	var printCenters []entity.PrintCenter
	if err := r.db.Find(&printCenters).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch all print centers: %w", err)
	}
	return printCenters, nil
}

func (r *printCenterRepository) Update(id uint, updates map[string]interface{}) error {
	result := r.db.Model(&entity.PrintCenter{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update print center id: %d: %w", id, result.Error)
	}
	return nil
}

func (r *printCenterRepository) Delete(id uint) error {
	result := r.db.Where("id = ?", id).Delete(&entity.PrintCenter{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete print center id: %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
