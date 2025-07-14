package service_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
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
}

func (s *UserServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mocks.NewMockUserRepository(s.ctrl)
	s.mockFbAuth = mocks.NewMockFirebaseAuthClient(s.ctrl)
	s.service = service.NewUserService(s.mockRepo, s.mockFbAuth)
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestRegister_Success() {
	user := &entity.User{
		UID:         "123",
		Role:        entity.RoleUser,
		Email:       "test@example.com",
		PhoneNumber: "+123456789",
	}

	s.mockRepo.EXPECT().FindByUID("123").Return(nil, gorm.ErrRecordNotFound)
	s.mockRepo.EXPECT().Save(gomock.Any()).Return(nil)

	result, err := s.service.Register(user)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(user.UID, result.UID)
}

func (s *UserServiceTestSuite) TestRegister_ExistingUser() {
	existingUser := &entity.User{UID: "123"}
	s.mockRepo.EXPECT().FindByUID("123").Return(existingUser, nil)

	_, err := s.service.Register(existingUser)

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserAlreadyExists)
}

func (s *UserServiceTestSuite) TestRegister_FindError() {
	dbErr := errors.New("database find error")
	s.mockRepo.EXPECT().FindByUID(gomock.Any()).Return(nil, dbErr)

	_, err := s.service.Register(&entity.User{UID: "123"})

	s.Error(err)
	s.ErrorContains(err, "failed to check for existing user")
}

func (s *UserServiceTestSuite) TestRegister_SaveError() {
	dbErr := errors.New("database save error")

	s.mockRepo.EXPECT().FindByUID(gomock.Any()).Return(nil, gorm.ErrRecordNotFound)
	s.mockRepo.EXPECT().Save(gomock.Any()).Return(dbErr)

	_, err := s.service.Register(&entity.User{UID: "123"})

	s.Error(err)
	s.ErrorContains(err, "failed to save new user")
}

func (s *UserServiceTestSuite) TestGetByUID_Success() {
	expectedUser := &entity.User{UID: "abc"}
	s.mockRepo.EXPECT().FindByUID("abc").Return(expectedUser, nil)

	user, err := s.service.GetByUID("abc")

	s.NoError(err)
	s.Equal(expectedUser, user)
}

func (s *UserServiceTestSuite) TestGetByUID_DBError() {
	dbErr := errors.New("database find error")
	s.mockRepo.EXPECT().FindByUID("abc").Return(nil, dbErr)

	_, err := s.service.GetByUID("abc")

	s.Error(err)
	s.ErrorContains(err, "getting user by UID")
}

func (s *UserServiceTestSuite) TestGetByUID_NotFound() {
	s.mockRepo.EXPECT().FindByUID("abc").Return(nil, gorm.ErrRecordNotFound)

	_, err := s.service.GetByUID("abc")

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestDelete_Success() {
	uid := "abc"
	s.mockFbAuth.EXPECT().DeleteUser(gomock.Any(), uid).Return(nil)
	s.mockRepo.EXPECT().Delete(uid).Return(nil)

	err := s.service.Delete(uid)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestDelete_FirebaseFails() {
	uid := "abc"
	fbErr := errors.New("internal firebase error")
	s.mockFbAuth.EXPECT().DeleteUser(gomock.Any(), uid).Return(fbErr)

	err := s.service.Delete(uid)

	s.Error(err)
	s.ErrorContains(err, "failed to delete user from firebase")
}

func (s *UserServiceTestSuite) TestDelete_FirebaseUserNotFound() {
	uid := "abc"
	fbErr := status.Error(codes.NotFound, "user not found")
	s.mockFbAuth.EXPECT().DeleteUser(gomock.Any(), uid).Return(fbErr)
	s.mockRepo.EXPECT().Delete(uid).Return(nil)

	err := s.service.Delete(uid)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestDelete_RepoFails() {
	uid := "abc"
	dbErr := errors.New("db delete error")
	s.mockFbAuth.EXPECT().DeleteUser(gomock.Any(), uid).Return(nil)
	s.mockRepo.EXPECT().Delete(uid).Return(dbErr)

	err := s.service.Delete(uid)

	s.Error(err)
	s.ErrorContains(err, "failed to delete user from database")
}

func (s *UserServiceTestSuite) TestDelete_RepoUserNotFound() {
	uid := "abc"
	s.mockFbAuth.EXPECT().DeleteUser(gomock.Any(), uid).Return(nil)
	s.mockRepo.EXPECT().Delete(uid).Return(gorm.ErrRecordNotFound)

	err := s.service.Delete(uid)

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestUpdateProfile_Success() {
	uid := "abc"
	existingUser := &entity.User{UID: uid, Email: "old@test.com"}
	updateRequest := &entity.User{UID: uid, Email: "new@test.com"}

	s.mockRepo.EXPECT().FindByUID(uid).Return(existingUser, nil)
	s.mockRepo.EXPECT().Update(uid, gomock.Any()).Return(nil)

	err := s.service.UpdateProfile(updateRequest)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestUpdateProfile_UpdateError() {
	uid := "abc"
	existingUser := &entity.User{UID: uid}
	updateRequest := &entity.User{UID: uid, Email: "new@test.com"}
	dbErr := errors.New("database update error")

	s.mockRepo.EXPECT().FindByUID(uid).Return(existingUser, nil)
	s.mockRepo.EXPECT().Update(uid, gomock.Any()).Return(dbErr)

	err := s.service.UpdateProfile(updateRequest)

	s.Error(err)
	s.ErrorContains(err, "updating user UID")
}

func (s *UserServiceTestSuite) TestUpdateProfile_UserNotFound() {
	s.mockRepo.EXPECT().FindByUID("abc").Return(nil, gorm.ErrRecordNotFound)

	err := s.service.UpdateProfile(&entity.User{UID: "abc"})

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestUpdateProfileByUID_Success() {
	uid := "abc"
	updates := map[string]interface{}{"phone_number": "+2291234556"}
	s.mockRepo.EXPECT().FindByUID(uid).Return(&entity.User{UID: uid}, nil)
	s.mockRepo.EXPECT().Update(uid, updates).Return(nil)

	err := s.service.UpdateProfileByUID(uid, updates)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestUpdateProfileByUID_UserNotFound() {
	uid := "abc"
	updates := map[string]any{"phone_number": "+2291234556"}
	s.mockRepo.EXPECT().FindByUID(uid).Return(nil, gorm.ErrRecordNotFound)

	err := s.service.UpdateProfileByUID(uid, updates)
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestUpdateProfileByUID_UpdateError() {
	uid := "abc"
	updates := map[string]any{"phone_number": "+2291234556"}
	dbErr := errors.New("database update error")

	s.mockRepo.EXPECT().FindByUID(uid).Return(&entity.User{UID: uid}, nil)
	s.mockRepo.EXPECT().Update(uid, updates).Return(dbErr)

	err := s.service.UpdateProfileByUID(uid, updates)

	s.Error(err)
	s.Equal(dbErr, err)
}

func (s *UserServiceTestSuite) TestGetAllUsers_Success() {
	expectedUsers := []entity.User{
		{UID: "u1"},
		{UID: "u2"},
	}
	s.mockRepo.EXPECT().FindAll().Return(expectedUsers, nil)

	users, err := s.service.GetAll()

	s.NoError(err)
	s.Len(users, 2)
	s.Equal("u1", users[0].UID)
}

func (s *UserServiceTestSuite) TestGetAll_Error() {
	dbErr := errors.New("database find all error")

	s.mockRepo.EXPECT().FindAll().Return(nil, dbErr)

	_, err := s.service.GetAll()

	s.Error(err)
	s.ErrorContains(err, "fetching all users")
}
