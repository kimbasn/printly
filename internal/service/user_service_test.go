package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/mocks"
	"github.com/kimbasn/printly/internal/service"
)

type UserServiceTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockRepo   *mocks.MockUserRepository
	mockFbAuth *mocks.MockFirebaseAuthClient
	service    service.UserService

	user         *entity.User
	password     string
	uid          string
	firebaseUser *auth.UserRecord
}

func (s *UserServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mocks.NewMockUserRepository(s.ctrl)
	s.mockFbAuth = mocks.NewMockFirebaseAuthClient(s.ctrl)
	s.service = service.NewUserService(s.mockRepo, s.mockFbAuth)

	// Given
	s.user = &entity.User{
		FirstName: "fname",
		LastName:  "lname",
		Role:      entity.RoleUser,
		Email:     "test@example.com",
	}
	s.password = "password123"
	s.uid = "firebase-uid-123"
	s.firebaseUser = &auth.UserRecord{UserInfo: &auth.UserInfo{UID: s.uid}}
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestRegister_Success() {
	// When
	s.mockFbAuth.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params *auth.UserToCreate) (*auth.UserRecord, error) {
			return s.firebaseUser, nil
		}).
		Times(1)

	s.mockRepo.EXPECT().
		Save(gomock.Any()).
		DoAndReturn(func(user *entity.User) error {
			// Verify user has been properly set up
			assert.Equal(s.T(), s.firebaseUser.UID, user.UID)
			assert.False(s.T(), user.CreatedAt.IsZero())
			assert.False(s.T(), user.UpdatedAt.IsZero())
			return nil
		}).
		Times(1)

	// Then
	result, err := s.service.Register(s.user, s.password)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(s.firebaseUser.UID, result.UID)
	s.Equal("test@example.com", result.Email)
	s.Equal("fname", result.FirstName)
	s.Equal("lname", result.LastName)
	s.False(result.CreatedAt.IsZero())
	s.False(result.UpdatedAt.IsZero())
	s.True(result.CreatedAt.Equal(result.UpdatedAt))
	s.WithinDuration(time.Now(), result.CreatedAt, time.Second)
}

func (s *UserServiceTestSuite) TestRegister_EmailAlreadyExists() {
	// When
	s.mockFbAuth.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(nil, ierrors.ErrEmailAlreadyExists).
		Times(1)

	// Then
	result, err := s.service.Register(s.user, s.password)

	// Assert
	s.Error(err)
	s.True(errors.Is(err, ierrors.ErrEmailAlreadyExists))
	s.Nil(result)
}

func (s *UserServiceTestSuite) TestRegister_FirebaseCreationFails() {
	// Given
	firebaseError := errors.New("firebase connection error")

	// When
	s.mockFbAuth.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(nil, firebaseError).
		Times(1)

	// Then
	result, err := s.service.Register(s.user, s.password)

	s.Error(err)
	s.Contains(err.Error(), "firebase connection error")
	s.Nil(result)
}

func (s *UserServiceTestSuite) TestRegister_DatabaseSaveFails_SuccessfulRollback() {
	dbError := errors.New("database connection error")

	// When
	s.mockFbAuth.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(s.firebaseUser, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Save(gomock.Any()).
		Return(dbError).
		Times(1)

	// Expect rollback call
	s.mockFbAuth.EXPECT().
		DeleteUser(gomock.Any(), s.firebaseUser.UID).
		Return(nil).
		Times(1)

	// Then
	result, err := s.service.Register(s.user, s.password)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "failed to save new user")
	s.Nil(result)
}

func (s *UserServiceTestSuite) TestRegister_DatabaseSaveFails_RollbackFails() {
	dbError := errors.New("database connection error")
	rollbackError := errors.New("firebase delete error")

	// When
	s.mockFbAuth.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(s.firebaseUser, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Save(gomock.Any()).
		Return(dbError).
		Times(1)

	// Expect rollback call that also fails
	s.mockFbAuth.EXPECT().
		DeleteUser(gomock.Any(), s.firebaseUser.UID).
		Return(rollbackError).Times(1)

	// Then
	result, err := s.service.Register(s.user, s.password)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "failed to save new user")
	s.Contains(err.Error(), "database connection error")
	s.Nil(result)

	// Note: The rollback error should be logged but not returned
}

func (s *UserServiceTestSuite) TestRegister_TimestampsSetCorrectly() {
	beforeCall := time.Now()

	// When
	s.mockFbAuth.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(s.firebaseUser, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Save(gomock.Any()).
		DoAndReturn(func(user *entity.User) error {
			afterCall := time.Now()

			// verify timestamps are within expected range
			assert.True(s.T(), user.CreatedAt.After(beforeCall) || user.CreatedAt.Equal(beforeCall))
			assert.True(s.T(), user.CreatedAt.Before(afterCall) || user.CreatedAt.Equal(afterCall))
			assert.True(s.T(), user.UpdatedAt.Equal(user.CreatedAt))

			return nil
		}).
		Times(1)

	// Then
	result, err := s.service.Register(s.user, s.password)

	// Assert
	s.NoError(err)
	s.NotNil(result)
}

func (s *UserServiceTestSuite) TestRegister_UserObjectModifiedCorrectly() {
	// When
	s.mockFbAuth.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(s.firebaseUser, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Save(gomock.Any()).
		Return(nil).
		Times(1)

	// Then
	result, err := s.service.Register(s.user, s.password)

	// Assert
	s.NoError(err)
	s.NotNil(result)

	// Verify the original user object was modified
	s.Equal(s.firebaseUser.UID, s.user.UID)
	s.False(s.user.CreatedAt.IsZero())
	s.False(s.user.UpdatedAt.IsZero())

	// Verify the returned user is the same object
	s.Equal(s.user, result)
}

// GetByUID Tests
func (s *UserServiceTestSuite) TestGetByUID_Success() {
	// Arrange
	uid := "test-uid"
	expectedUser := &entity.User{
		UID:       uid,
		FirstName: "fname",
		LastName:  "lname",
		Role:      entity.RoleUser,
	}
	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(expectedUser, nil)

	// Act
	user, err := s.service.GetByUID(uid)

	// Assert
	s.NoError(err)
	s.Equal(expectedUser, user)
}

func (s *UserServiceTestSuite) TestGetByUID_UserNotFound() {
	// Arrange
	uid := "non-existent-uid"
	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	user, err := s.service.GetByUID(uid)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrUserNotFound, err)
	s.Nil(user)
}

func (s *UserServiceTestSuite) TestGetByUID_DatabaseError() {
	// Arrange
	uid := "test-uid"
	dbError := errors.New("database connection error")
	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(nil, dbError)

	// Act
	user, err := s.service.GetByUID(uid)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "getting user by UID test-uid")
	s.Nil(user)
}

func (s *UserServiceTestSuite) TestGetByUID_EmptyUID() {
	// Arrange
	uid := ""
	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	user, err := s.service.GetByUID(uid)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrUserNotFound, err)
	s.Nil(user)
}

// UpdateProfile Tests
func (s *UserServiceTestSuite) TestUpdateProfile_Success() {
	// Arrange
	uid := "test-uid"
	updates := map[string]any{
		"first_name": "Updated First",
		"last_name":  "Updated Last",
		"email":      "updated@example.com",
	}
	existingUser := &entity.User{
		UID:       uid,
		FirstName: "Old First",
		LastName:  "Old Last",
		Email:     "old@example.com",
	}

	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(existingUser, nil)

	s.mockRepo.EXPECT().
		Update(uid, updates).
		Return(nil)

	// Act
	err := s.service.UpdateProfile(uid, updates)

	// Assert
	s.NoError(err)
}

func (s *UserServiceTestSuite) TestUpdateProfile_UserNotFound() {
	// Arrange
	uid := "non-existent-uid"
	updates := map[string]any{"first_name": "Updated Name"}
	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	err := s.service.UpdateProfile(uid, updates)

	// Assert
	s.Error(err)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateProfile_UpdateError() {
	// Arrange
	uid := "test-uid"
	updates := map[string]any{"first_name": "Updated Name"}
	existingUser := &entity.User{
		UID:       uid,
		FirstName: "Old Name",
	}
	updateError := errors.New("update failed")

	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(existingUser, nil)
	s.mockRepo.EXPECT().
		Update(uid, updates).
		Return(updateError)

	// Act
	err := s.service.UpdateProfile(uid, updates)

	// Assert
	s.Error(err)
	s.Equal(updateError, err)
}

func (s *UserServiceTestSuite) TestUpdateProfile_EmptyUpdates() {
	// Arrange
	uid := "test-uid"
	updates := map[string]any{}
	existingUser := &entity.User{
		UID:       uid,
		FirstName: "Test User",
	}

	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(existingUser, nil)
	s.mockRepo.EXPECT().
		Update(uid, updates).
		Return(nil)

	// Act
	err := s.service.UpdateProfile(uid, updates)

	// Assert
	s.NoError(err)
}

func (s *UserServiceTestSuite) TestUpdateRole_Success() {
	// Arrange
	uid := "test-uid"
	role := entity.RoleAdmin
	existingUser := &entity.User{
		UID:  uid,
		Role: entity.RoleUser,
	}
	expectedUpdates := map[string]any{"role": role}
	updatedUser := &entity.User{
		UID:  uid,
		Role: role,
	}

	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(existingUser, nil)
	s.mockRepo.EXPECT().
		Update(uid, expectedUpdates).
		Return(nil)
	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(updatedUser, nil)

	// Act
	err := s.service.UpdateRole(uid, role)
	s.NoError(err)

	user, err := s.service.GetByUID(uid)

	s.NoError(err)
	s.Equal(user.Role, role)
}

func (s *UserServiceTestSuite) TestUpdateRole_UserNotFound() {
	// Arrange
	uid := "non-existent-uid"
	role := entity.RoleAdmin
	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(nil, ierrors.ErrUserNotFound)

	// Act
	err := s.service.UpdateRole(uid, role)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateRole_UpdateError() {
	// Arrange
	uid := "test-uid"
	role := entity.RoleAdmin
	existingUser := &entity.User{
		UID:  uid,
		Role: entity.RoleUser,
	}
	expectedUpdates := map[string]any{"role": role}
	updateError := errors.New("role update failed")

	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(existingUser, nil)
	s.mockRepo.EXPECT().
		Update(uid, expectedUpdates).
		Return(updateError)

	// Act
	err := s.service.UpdateRole(uid, role)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "updating role for user UID test-uid")
}

func (s *UserServiceTestSuite) TestUpdateRole_SameRole() {
	// Arrange
	uid := "test-uid"
	role := entity.RoleUser
	existingUser := &entity.User{
		UID:  uid,
		Role: role,
	}
	expectedUpdates := map[string]any{"role": role}

	s.mockRepo.EXPECT().
		FindByUID(uid).
		Return(existingUser, nil)
	s.mockRepo.EXPECT().
		Update(uid, expectedUpdates).
		Return(nil)

	// Act
	err := s.service.UpdateRole(uid, role)
	s.NoError(err)
}


// Delete Tests
func (s *UserServiceTestSuite) TestDelete_Success() {
	// Arrange
	uid := "test-uid"
	s.mockFbAuth.EXPECT().
		DeleteUser(gomock.Any(), uid).
		Return(nil)
	s.mockRepo.EXPECT().
		Delete(uid).
		Return(nil)

	// Act
	err := s.service.Delete(uid)

	// Assert
	s.NoError(err)
}

func (s *UserServiceTestSuite) TestDelete_FirebaseUserNotFound_Success() {
	// Arrange
	uid := "test-uid"
	firebaseNotFoundError := status.Error(codes.NotFound, "user not found")
	s.mockFbAuth.EXPECT().
		DeleteUser(gomock.Any(), uid).
		Return(firebaseNotFoundError)
	s.mockRepo.EXPECT().
		Delete(uid).
		Return(nil)

	// Act
	err := s.service.Delete(uid)

	// Assert
	s.NoError(err)
}

func (s *UserServiceTestSuite) TestDelete_FirebaseError() {
	// Arrange
	uid := "test-uid"
	firebaseError := status.Error(codes.Internal, "internal firebase error")
	s.mockFbAuth.EXPECT().
		DeleteUser(gomock.Any(), uid).
		Return(firebaseError)

	// Act
	err := s.service.Delete(uid)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "failed to delete user from firebase UID test-uid")
}

func (s *UserServiceTestSuite) TestDelete_DatabaseUserNotFound() {
	// Arrange
	uid := "test-uid"
	s.mockFbAuth.EXPECT().
		DeleteUser(gomock.Any(), uid).
		Return(nil)
	s.mockRepo.EXPECT().
		Delete(uid).
		Return(gorm.ErrRecordNotFound)

	// Act
	err := s.service.Delete(uid)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestDelete_DatabaseError() {
	// Arrange
	uid := "test-uid"
	dbError := errors.New("database deletion error")
	s.mockFbAuth.EXPECT().
		DeleteUser(gomock.Any(), uid).
		Return(nil)
	s.mockRepo.EXPECT().
		Delete(uid).
		Return(dbError)

	// Act
	err := s.service.Delete(uid)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "failed to delete user from database UID test-uid")
}

// GetAll Tests
func (s *UserServiceTestSuite) TestGetAll_Success() {
	// Arrange
	expectedUsers := []entity.User{
		{
			UID:       "uid1",
			FirstName: "User",
			LastName:  "One",
			Role:      entity.RoleUser,
		},
		{
			UID:       "uid2",
			FirstName: "User",
			LastName:  "Two",
			Role:      entity.RoleAdmin,
		},
	}
	s.mockRepo.EXPECT().
		FindAll().
		Return(expectedUsers, nil)

	// Act
	users, err := s.service.GetAll()

	// Assert
	s.NoError(err)
	s.Equal(expectedUsers, users)
	s.Len(users, 2)
}

func (s *UserServiceTestSuite) TestGetAll_EmptyResult() {
	// Arrange
	expectedUsers := []entity.User{}
	s.mockRepo.EXPECT().
		FindAll().
		Return(expectedUsers, nil)

	// Act
	users, err := s.service.GetAll()

	// Assert
	s.NoError(err)
	s.Equal(expectedUsers, users)
	s.Len(users, 0)
}

func (s *UserServiceTestSuite) TestGetAll_DatabaseError() {
	// Arrange
	dbError := errors.New("database query error")
	s.mockRepo.EXPECT().
		FindAll().
		Return(nil, dbError)

	// Act
	users, err := s.service.GetAll()

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "fetching all users")
	s.Nil(users)
}
