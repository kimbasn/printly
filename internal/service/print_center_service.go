package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/repository"
	"gorm.io/gorm"
)

// PrintCenterService defines the interface for print center-related business logic.
type PrintCenterService interface {
	Register(center *entity.PrintCenter) (*entity.PrintCenter, error)
	GetByID(id uint) (*entity.PrintCenter, error)
	GetApproved() ([]entity.PrintCenter, error)
	GetPending() ([]entity.PrintCenter, error)
	GetAll() ([]entity.PrintCenter, error)
	Update(id uint, updates map[string]interface{}) error
	UpdateStatus(id uint, status entity.PrintCenterStatus) error
	Delete(id uint) error
}

type printCenterService struct {
	repo repository.PrintCenterRepository
}

// NewPrintCenterService creates a new instance of PrintCenterService.
func NewPrintCenterService(repo repository.PrintCenterRepository) PrintCenterService {
	return &printCenterService{repo: repo}
}

// Register creates a new print center. It's initially set to 'pending' status.
func (s *printCenterService) Register(center *entity.PrintCenter) (*entity.PrintCenter, error) {
	// later need to check uniqueness of address or geographical coordiantes before saving

	newCenter := *center

	now := time.Now()
	newCenter.CreatedAt = now
	newCenter.UpdatedAt = now
	newCenter.Status = entity.StatusPending // Set default status

	if err := s.repo.Save(&newCenter); err != nil {
		return nil, fmt.Errorf("failed to save new print center: %w", err)
	}
	return &newCenter, nil
}

// GetByID retrieves a print center by its ID.
func (s *printCenterService) GetByID(id uint) (*entity.PrintCenter, error) {
	center, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ierrors.ErrPrintCenterNotFound
		}
		return nil, fmt.Errorf("getting print center by id %d: %w", id, err)
	}

	return center, nil
}

// GetAllPublic retrieves all approved print centers.
func (s *printCenterService) GetApproved() ([]entity.PrintCenter, error) {
	centers, err := s.repo.FindByStatus(entity.StatusApproved)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch approved print centers: %w", err)
	}
	return centers, nil
}

// GetPending retrieves all print centers awaiting approval.
func (s *printCenterService) GetPending() ([]entity.PrintCenter, error) {
	centers, err := s.repo.FindByStatus(entity.StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending print centers: %w", err)
	}
	return centers, nil
}

// GetAll retrieves all print centers, regardless of status (for admin use).
func (s *printCenterService) GetAll() ([]entity.PrintCenter, error) {
	centers, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all print centers: %w", err)
	}
	return centers, nil
}

// Update performs a partial update on a print center's properties.
func (s *printCenterService) Update(id uint, updates map[string]any) error {
	if _, err := s.GetByID(id); err != nil {
		return err // Will be ErrPrintCenterNotFound or a db error
	}

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()
	if err := s.repo.Update(id, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ierrors.ErrPrintCenterNotFound
		}
		return fmt.Errorf("updating print center id %d: %w", id, err)
	}

	return nil
}

// UpdateStatus updates the status of a print center (e.g., approve, suspend).
func (s *printCenterService) UpdateStatus(id uint, status entity.PrintCenterStatus) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}

	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
	}

	return s.repo.Update(id, updates)
}

// Delete removes a print center from the database.
func (s *printCenterService) Delete(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ierrors.ErrPrintCenterNotFound
		}
		return fmt.Errorf("failed to delete print center id %d: %w", id, err)
	}
	return nil
}
