package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/mocks"
	"github.com/kimbasn/printly/internal/service"
)

type UserServiceTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mocks.MockUserRepository
	service  service.UserService
}

func (s *UserServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mocks.NewMockUserRepository(s.ctrl)
	s.service = service.NewUserService(s.mockRepo)
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestRegister_NewUser() {
	user := &entity.User{
		UID:         "123",
		Role:        entity.RoleUser,
		Email:       "test@example.com",
		PhoneNumber: "+123456789",
	}

	s.mockRepo.EXPECT().FindByUID("123").Return(nil, nil)
	// Use gomock.Any() because the service modifies the user object (timestamps) before saving.
	s.mockRepo.EXPECT().Save(gomock.Any()).Return(nil)

	result, err := s.service.Register(user)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(user.UID, result.UID)
	s.WithinDuration(time.Now(), result.CreatedAt, time.Second)
	s.WithinDuration(time.Now(), result.UpdatedAt, time.Second)
}

func (s *UserServiceTestSuite) TestRegister_ExistingUser() {
	existingUser := &entity.User{UID: "123"}
	s.mockRepo.EXPECT().FindByUID("123").Return(existingUser, nil)

	_, err := s.service.Register(&entity.User{UID: "123"})

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserAlreadyExists)
}

func (s *UserServiceTestSuite) TestRegister_SaveError() {
	dbErr := errors.New("database save error")

	s.mockRepo.EXPECT().FindByUID(gomock.Any()).Return(nil, nil)
	s.mockRepo.EXPECT().Save(gomock.Any()).Return(dbErr)

	_, err := s.service.Register(&entity.User{UID: "123"})

	s.Error(err)
	s.ErrorContains(err, dbErr.Error())
}

func (s *UserServiceTestSuite) TestGetByUID_Success() {
	expectedUser := &entity.User{UID: "abc"}
	s.mockRepo.EXPECT().FindByUID("abc").Return(expectedUser, nil)

	result, err := s.service.GetByUID("abc")

	s.NoError(err)
	s.Equal(expectedUser, result)
}

func (s *UserServiceTestSuite) TestGetByUID_NotFound() {
	s.mockRepo.EXPECT().FindByUID("abc").Return(nil, nil)

	_, err := s.service.GetByUID("abc")

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestDeleteUser_Success() {
	s.mockRepo.EXPECT().Delete("abc").Return(nil)

	err := s.service.Delete("abc")

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestDeleteUser_NotFound() {
	s.mockRepo.EXPECT().Delete("abc").Return(gorm.ErrRecordNotFound)

	err := s.service.Delete("abc")

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestUpdateProfile_Success() {
	existingUser := &entity.User{UID: "abc", Email: "old@test.com"}
	updateRequest := &entity.User{UID: "abc", Email: "new@test.com"}

	s.mockRepo.EXPECT().FindByUID("abc").Return(existingUser, nil)
	s.mockRepo.EXPECT().Update("abc", gomock.Any()).DoAndReturn(func(_ string, updates map[string]interface{}) error {
		s.Equal("new@test.com", updates["email"])
		s.WithinDuration(time.Now(), updates["updated_at"].(time.Time), time.Second)
		return nil
	})

	err := s.service.UpdateProfile(updateRequest)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestUpdateProfile_NotFound() {
	s.mockRepo.EXPECT().FindByUID("abc").Return(nil, nil)

	err := s.service.UpdateProfile(&entity.User{UID: "abc"})

	s.Error(err)
	s.ErrorIs(err, ierrors.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestGetAllUsers_Success() {
	expectedUsers := []entity.User{
		{UID: "u1"},
		{UID: "u2"},
	}
	s.mockRepo.EXPECT().FindAll().Return(expectedUsers, nil)

	result, err := s.service.GetAll()

	s.NoError(err)
	s.Len(result, 2)
	s.Equal("u1", result[0].UID)
}

func (s *UserServiceTestSuite) TestGetAllUsers_Error() {
	dbErr := errors.New("database find all error")

	s.mockRepo.EXPECT().FindAll().Return(nil, dbErr)

	_, err := s.service.GetAll()

	s.Error(err)
	s.ErrorContains(err, dbErr.Error())
}
