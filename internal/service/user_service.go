package service

import (
	"fmt"
	"time"
	"errors"

	"github.com/kimbasn/printly/internal/entity"
	ierrors"github.com/kimbasn/printly/internal/errors"

	"github.com/kimbasn/printly/internal/repository"
	"gorm.io/gorm"
)

type UserService interface {
	Register(user *entity.User) (*entity.User, error)
	GetByUID(uid string) (*entity.User, error)
	Delete(uid string) error
	UpdateProfile(user *entity.User) error
	GetAll() ([]entity.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(user *entity.User) (*entity.User, error) {
	existing, err := s.repo.FindByUID(user.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if existing != nil {
		return nil, ierrors.ErrUserAlreadyExists
	}

	// Set timestamps for new user
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := s.repo.Save(user); err != nil {
		return nil, fmt.Errorf("failed to save new user: %w", err)
	}
	return user, nil
}

func (s *userService) GetByUID(uid string) (*entity.User, error) {
	user, err := s.repo.FindByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("getting user by UID %s: %w", uid, err)
	}
	if user == nil {
		return nil, ierrors.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) UpdateProfile(user *entity.User) error {
	existing, err := s.repo.FindByUID(user.UID)
	if err != nil {
		return fmt.Errorf("failed to find user for update: %w", err)
	}
	if existing == nil {
		return ierrors.ErrUserNotFound
	}

	// Build a map of fields to update to perform a partial update.
	// This is more efficient and prevents accidental updates to protected fields.
	updates := make(map[string]interface{})
	if user.Email != "" {
		updates["email"] = user.Email
	}
	if user.PhoneNumber != "" {
		updates["phone_number"] = user.PhoneNumber
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.repo.Update(user.UID, updates); err != nil {
			return fmt.Errorf("updating user UID %s: %w", user.UID, err)
		}
	}

	return nil
}

func (s *userService) Delete(uid string) error {
	err := s.repo.Delete(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ierrors.ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user UID %s: %w", uid, err)
	}
	return nil
}

func (s *userService) GetAll() ([]entity.User, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("fetching all users: %w", err)
	}
	return users, nil
}