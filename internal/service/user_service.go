package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"

	"github.com/kimbasn/printly/internal/repository"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../mocks/mock_firebase_auth_client.go -package=mocks github.com/kimbasn/printly/internal/service FirebaseAuthClient

// FirebaseAuthClient defines an interface for Firebase Auth operations, allowing for mocking.
type FirebaseAuthClient interface {
	DeleteUser(ctx context.Context, uid string) error
}

type UserService interface {
	Register(user *entity.User) (*entity.User, error)
	GetByUID(uid string) (*entity.User, error)
	Delete(uid string) error
	UpdateProfile(user *entity.User) error
	GetAll() ([]entity.User, error)
	UpdateProfileByUID(uid string, updates map[string]any) error
}

type userService struct {
	repo   repository.UserRepository
	fbAuth FirebaseAuthClient
}

func NewUserService(repo repository.UserRepository, fbAuth FirebaseAuthClient) UserService {
	return &userService{
		repo:   repo,
		fbAuth: fbAuth,
	}
}

func (s *userService) Register(user *entity.User) (*entity.User, error) {
	_, err := s.repo.FindByUID(user.UID)
	if err == nil {
		return nil, ierrors.ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ierrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("getting user by UID %s: %w", uid, err)
	}
	return user, nil
}

func (s *userService) UpdateProfile(user *entity.User) error {
	// First, ensure the user exists. GetByUID already handles not found errors.
	if _, err := s.GetByUID(user.UID); err != nil {
		return err
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

func (s *userService) UpdateProfileByUID(uid string, updates map[string]any) error {
	// Ensure the user exists before updating.
	if _, err := s.GetByUID(uid); err != nil {
		return err
	}
	return s.repo.Update(uid, updates)
}

func (s *userService) Delete(uid string) error {
	// First, delete the user from Firebase Authentication.
	if err := s.fbAuth.DeleteUser(context.Background(), uid); err != nil {
		// If the user is not found in Firebase, that's okay. We can proceed to delete them from our DB.
		// For any other Firebase error, we should stop to avoid data inconsistency.
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("failed to delete user from firebase UID %s: %w", uid, err)
		}
		log.Printf("User with UID %s not found in Firebase, proceeding with local DB deletion.", uid)
	}
	// Then, delete the user from the local database
	if err := s.repo.Delete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ierrors.ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user from database UID %s: %w", uid, err)
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
