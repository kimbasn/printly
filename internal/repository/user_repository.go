package repository

import (
	"errors"
	"fmt"

	"github.com/kimbasn/printly/internal/entity"

	"gorm.io/gorm"
)

//go:generate mockgen -destination=../mocks/mock_user_repository.go -package=mocks github.com/kimbasn/printly/internal/repository UserRepository

// UserRepository defines the interface for user-related database operations.
type UserRepository interface {
	Save(user *entity.User) error
	FindByUID(uid string) (*entity.User, error)
	Delete(uid string) error
	Update(uid string, updates map[string]interface{}) error
	FindAll() ([]entity.User, error)
}

// userRepository implements the UserRepository interface using GORM.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of a UserRepository.
// It takes a GORM database connection as a dependency.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Save creates a new user record in the database.
// It returns an error if the database operation fails.
func (r *userRepository) Save(u *entity.User) error {
	if err := r.db.Create(u).Error; err != nil {
		return fmt.Errorf("failed to save the user: %w", err)
	}
	return nil
}

// FindByUID retrieves a user from the database by their unique identifier (UID).
// If no user is found, it returns (nil, nil).
// It returns an error for any other database-related issues.
func (r *userRepository) FindByUID(uid string) (*entity.User, error) {
	var user entity.User
	result := r.db.First(&user, "uid = ?", uid)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch user UID %s: %w", uid, result.Error)
	}
	return &user, nil
}

// Update modifies an existing user's record in the database.
// It uses GORM's Save method, which updates all fields of the user struct.
func (r *userRepository) Update(uid string, update map[string]interface{}) error {
	result := r.db.Model(&entity.User{}).Where("uid = ?", uid).Updates(update)
	if result.Error != nil {
		return fmt.Errorf("failed to update user UID %s: %w", uid, result.Error)
	}
	return nil
}

// Delete removes a user from the database based on their UID.
func (r *userRepository) Delete(uid string) error {
	result := r.db.Where("uid = ?", uid).Delete(&entity.User{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete user UID %s: %w", uid, result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindAll retrieves all user records from the database.
func (r *userRepository) FindAll() ([]entity.User, error) {
	var users []entity.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch all users: %w", err)
	}
	return users, nil
}
