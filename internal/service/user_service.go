package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"firebase.google.com/go/v4/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kimbasn/printly/internal/adapter"
	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"

	"github.com/kimbasn/printly/internal/repository"
	"gorm.io/gorm"
)

type UserService interface {
	Register(user *entity.User, password string) (*entity.User, error)
	GetByUID(uid string) (*entity.User, error)
	Delete(uid string) error
	GetAll() ([]entity.User, error)
	UpdateProfile(uid string, updates map[string]any) error
	UpdateRole(uid string, role entity.Role) error
}

type userService struct {
	repo   repository.UserRepository
	fbAuth adapter.FirebaseAuthClient
}

func NewUserService(repo repository.UserRepository, fbAuth adapter.FirebaseAuthClient) UserService {
	return &userService{
		repo:   repo,
		fbAuth: fbAuth,
	}
}

func (s *userService) Register(user *entity.User, password string) (*entity.User, error) {
	// 1. Create user in Firebase
	params := (&auth.UserToCreate{}).
		DisplayName(user.FirstName + " " + user.LastName).
		Email(user.Email).
		Password(password).
		Disabled(false)

	fbUser, err := s.fbAuth.CreateUser(context.Background(), params)
	if err != nil {
		return nil, err
	}

	// 2. User created in firebase, now save to local DB.
	// Set the UID from Firebase response
	user.UID = fbUser.UID

	// Set timestamps for new user
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := s.repo.Save(user); err != nil {
		log.Printf("CRITICAL: Failed to save user %s to DB after Firebase creation. Attempting rollback", fbUser.UID)
		if rollbackErr := s.fbAuth.DeleteUser(context.Background(), fbUser.UID); rollbackErr != nil {
			log.Printf("CRITICAL: FAILED TO ROLLBACK FIREBASE USER %s. MANUAL INTERVENTION REQUIRED. Error: %v", fbUser.UID, rollbackErr)
		}
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

func (s *userService) UpdateProfile(uid string, updates map[string]any) error {
	// Ensure the user exists before updating.
	if _, err := s.repo.FindByUID(uid); err != nil {
		return err
	}
	return s.repo.Update(uid, updates)
}

func (s *userService) UpdateRole(uid string, role entity.Role) error {
	// 1. Ensure the user exists
	if _, err := s.GetByUID(uid); err != nil {
		return err
	}

	// 2. Update the role in the local database
	updates := map[string]any{"role": role}
	if err := s.repo.Update(uid, updates); err != nil {
		return fmt.Errorf("updating role for user UID %s: %w", uid, err)
	}

	// 3. Set custom claims in firebase
	// claims := map[string]any{"role": string(role)}
	// if err := s.fbAuth.SetCustomerClaims(context.Background(), uid, claims); err != nil {
	// 	return fmt.Errorf("failed to set custom claims in firebase for UID %s: %w", uid, err)
	// }
	return nil
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
