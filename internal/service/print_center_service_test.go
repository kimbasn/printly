package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/mocks"
	"github.com/kimbasn/printly/internal/service"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type PrintCenterServiceTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mocks.MockPrintCenterRepository
	service  service.PrintCenterService
}

func (s *PrintCenterServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mocks.NewMockPrintCenterRepository(s.ctrl)
	s.service = service.NewPrintCenterService(s.mockRepo)
}

func (s *PrintCenterServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestPrintCenterService(t *testing.T) {
	suite.Run(t, new(PrintCenterServiceTestSuite))
}

// ============================================================================
// Register Tests
// ============================================================================
func (s *PrintCenterServiceTestSuite) TestRegister_Success() {
	// Arrange
	center := &entity.PrintCenter{Name: "New Center"}
	s.mockRepo.EXPECT().Save(gomock.Any()).DoAndReturn(func(c *entity.PrintCenter) error {
		s.Equal(entity.StatusPending, c.Status)
		s.WithinDuration(time.Now(), c.CreatedAt, time.Second)
		return nil
	})

	// Act
	result, err := s.service.Register(center)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal("New Center", result.Name)
}

func (s *PrintCenterServiceTestSuite) TestRegister_SaveError() {
	// Arrange
	dbErr := errors.New("db save error")
	s.mockRepo.EXPECT().Save(gomock.Any()).Return(dbErr)

	// Act
	_, err := s.service.Register(&entity.PrintCenter{})

	// Assert
	s.Error(err)
	s.ErrorContains(err, dbErr.Error())
}

func (s *PrintCenterServiceTestSuite) TestGetByID_Success() {
	// Arrange
	var centerID uint = 1
	expectedCenter := &entity.PrintCenter{ID: 1, Name: "Test Center"}
	s.mockRepo.EXPECT().FindByID(centerID).Return(expectedCenter, nil)

	// Act
	result, err := s.service.GetByID(centerID)

	// Assert
	s.NoError(err)
	s.Equal(expectedCenter, result)
}

func (s *PrintCenterServiceTestSuite) TestGetByID_NotFound() {
	// Arrange
	var centerID uint = 1
	s.mockRepo.EXPECT().FindByID(centerID).Return(nil, nil)

	// Act
	_, err := s.service.GetByID(centerID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrPrintCenterNotFound)
}

func (s *PrintCenterServiceTestSuite) TestGetAllPublic_Success() {
	// Arrange
	expectedCenters := []entity.PrintCenter{{Name: "Approved Center"}}
	s.mockRepo.EXPECT().FindByStatus(entity.StatusApproved).Return(expectedCenters, nil)

	// Act
	result, err := s.service.GetApproved()

	// Assert
	s.NoError(err)
	s.Equal(expectedCenters, result)
}

func (s *PrintCenterServiceTestSuite) TestUpdate_Success() {
	// Arrange
	var centerID uint = 1
	updates := map[string]interface{}{"name": "Updated Name"}
	existingCenter := &entity.PrintCenter{ID: 1}

	s.mockRepo.EXPECT().FindByID(centerID).Return(existingCenter, nil)
	s.mockRepo.EXPECT().Update(centerID, gomock.Any()).Return(nil)

	// Act
	err := s.service.Update(centerID, updates)

	// Assert
	s.NoError(err)
}

func (s *PrintCenterServiceTestSuite) TestUpdate_NotFound() {
	// Arrange
	var centerID uint = 1
	updates := map[string]any{"name": "Updated Name"}
	s.mockRepo.EXPECT().FindByID(centerID).Return(nil, ierrors.ErrPrintCenterNotFound)

	// Act
	err := s.service.Update(centerID, updates)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrPrintCenterNotFound)
}

func (s *PrintCenterServiceTestSuite) TestDelete_Success() {
	// Arrange
	var centerID uint = 1
	s.mockRepo.EXPECT().Delete(centerID).Return(nil)

	// Act
	err := s.service.Delete(centerID)

	// Assert
	s.NoError(err)
}

func (s *PrintCenterServiceTestSuite) TestDelete_NotFound() {
	// Arrange
	var centerID uint = 1
	s.mockRepo.EXPECT().Delete(centerID).Return(gorm.ErrRecordNotFound)

	// Act
	err := s.service.Delete(centerID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrPrintCenterNotFound)
}

