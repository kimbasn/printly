package service_test

import (
	"errors"
	"fmt"
	"strconv"
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
	_, err := s.service.Register(&entity.PrintCenter{Name: "Test Center"})

	// Assert
	s.Error(err)
	s.ErrorContains(err, "failed to save new print center")
	s.ErrorContains(err, dbErr.Error())
}

// ============================================================================
// GetByID Tests
// ============================================================================

func (s *PrintCenterServiceTestSuite) TestGetByID_Success() {
	// Arrange
	var centerID uint = 1
	expectedCenter := &entity.PrintCenter{
		ID:   1,
		Name: "Test Center",
		Address: entity.Address{
			Number: "1",
			Type:   "Avenue",
			Street: "Kimba SABI N'GOYE",
			City:   "Kandi",
		},
		Status: entity.StatusApproved,
	}
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
	s.mockRepo.EXPECT().FindByID(centerID).Return(nil, gorm.ErrRecordNotFound)

	// Act
	_, err := s.service.GetByID(centerID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrPrintCenterNotFound)
}

func (s *PrintCenterServiceTestSuite) TestGetByID_DatabaseError() {
	// Arrange
	var centerID uint = 1
	dbErr := errors.New("database connection error")
	s.mockRepo.EXPECT().FindByID(centerID).Return(nil, dbErr)

	// Act
	_, err := s.service.GetByID(centerID)

	// Assert
	s.Error(err)
	s.ErrorContains(err, "getting print center by id 1")
	s.ErrorContains(err, dbErr.Error())
}

// ============================================================================
// GetApproved Tests
// ============================================================================

func (s *PrintCenterServiceTestSuite) TestGetApproved_Success() {
	// Arrange
	expectedCenters := []entity.PrintCenter{
		{ID: 1, Name: "Center 1", Status: entity.StatusApproved},
		{ID: 2, Name: "Center 2", Status: entity.StatusApproved},
	}
	s.mockRepo.EXPECT().FindByStatus(entity.StatusApproved).Return(expectedCenters, nil)

	// Act
	result, err := s.service.GetApproved()

	// Assert
	s.NoError(err)
	s.Equal(expectedCenters, result)
	s.Len(result, 2)
}

func (s *PrintCenterServiceTestSuite) TestGetApproved_EmptyResult() {
	// Arrange
	s.mockRepo.EXPECT().FindByStatus(entity.StatusApproved).Return([]entity.PrintCenter{}, nil)

	// Act
	result, err := s.service.GetApproved()

	// Assert
	s.NoError(err)
	s.Empty(result)
}

func (s *PrintCenterServiceTestSuite) TestGetApproved_DatabaseError() {
	// Arrange
	dbErr := errors.New("database error")
	s.mockRepo.EXPECT().FindByStatus(entity.StatusApproved).Return(nil, dbErr)

	// Act
	_, err := s.service.GetApproved()

	// Assert
	s.Error(err)
	s.ErrorContains(err, "failed to fetch approved print centers")
}

// ============================================================================
// GetPending Tests
// ============================================================================

func (s *PrintCenterServiceTestSuite) TestGetPending_Success() {
	// Arrange
	expectedCenters := []entity.PrintCenter{
		{ID: 1, Name: "Pending Center 1", Status: entity.StatusPending},
		{ID: 2, Name: "Pending Center 2", Status: entity.StatusPending},
	}
	s.mockRepo.EXPECT().FindByStatus(entity.StatusPending).Return(expectedCenters, nil)

	// Act
	result, err := s.service.GetPending()

	// Assert
	s.NoError(err)
	s.Equal(expectedCenters, result)
}

func (s *PrintCenterServiceTestSuite) TestGetPending_DatabaseError() {
	// Arrange
	dbErr := errors.New("database error")
	s.mockRepo.EXPECT().FindByStatus(entity.StatusPending).Return(nil, dbErr)

	// Act
	_, err := s.service.GetPending()

	// Assert
	s.Error(err)
	s.ErrorContains(err, "failed to fetch pending print centers")
}

// ============================================================================
// GetAll Tests
// ============================================================================

func (s *PrintCenterServiceTestSuite) TestGetAll_Success() {
	// Arrange
	expectedCenters := []entity.PrintCenter{
		{ID: 1, Name: "Center 1", Status: entity.StatusApproved},
		{ID: 2, Name: "Center 2", Status: entity.StatusPending},
		{ID: 3, Name: "Center 3", Status: entity.StatusSuspended},
	}
	s.mockRepo.EXPECT().FindAll().Return(expectedCenters, nil)

	// Act
	result, err := s.service.GetAll()

	// Assert
	s.NoError(err)
	s.Equal(expectedCenters, result)
	s.Len(result, 3)
}

func (s *PrintCenterServiceTestSuite) TestGetAll_DatabaseError() {
	// Arrange
	dbErr := errors.New("database error")
	s.mockRepo.EXPECT().FindAll().Return(nil, dbErr)

	// Act
	_, err := s.service.GetAll()

	// Assert
	s.Error(err)
	s.ErrorContains(err, "failed to fetch all print centers")
}

// ============================================================================
// Update Tests
// ============================================================================

func (s *PrintCenterServiceTestSuite) TestUpdate_Success() {
	// Arrange
	var centerID uint = 1
	updates := map[string]any{"name": "Updated Name"}
	existingCenter := &entity.PrintCenter{ID: 1, Name: "Original Name"}

	s.mockRepo.EXPECT().FindByID(centerID).Return(existingCenter, nil)
	s.mockRepo.EXPECT().Update(centerID, gomock.Any()).DoAndReturn(func(id uint, updates map[string]any) error {
		s.Contains(updates, "updated_at")
		s.Contains(updates, "name")
		s.Equal("Updated Name", updates["name"])
		return nil
	})

	// Act
	err := s.service.Update(centerID, updates)

	// Assert
	s.NoError(err)
}

func (s *PrintCenterServiceTestSuite) TestUpdate_NotFound() {
	// Arrange
	var centerID uint = 1
	updates := map[string]any{"name": "Updated Name"}
	s.mockRepo.EXPECT().FindByID(centerID).Return(nil, gorm.ErrRecordNotFound)

	// Act
	err := s.service.Update(centerID, updates)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrPrintCenterNotFound)
}

func (s *PrintCenterServiceTestSuite) TestUpdate_EmptyUpdates() {
	// Arrange
	var centerID uint = 1
	existingCenter := &entity.PrintCenter{ID: 1}

	s.mockRepo.EXPECT().FindByID(centerID).Return(existingCenter, nil)
	// No Update call should be made for empty updates

	// Act
	err := s.service.Update(centerID, map[string]interface{}{})

	// Assert
	s.NoError(err)
}

func (s *PrintCenterServiceTestSuite) TestUpdate_DatabaseError() {
	// Arrange
	var centerID uint = 1
	updates := map[string]any{"name": "Updated Name"}
	existingCenter := &entity.PrintCenter{ID: 1}
	dbErr := errors.New("database error")

	s.mockRepo.EXPECT().FindByID(centerID).Return(existingCenter, nil)
	s.mockRepo.EXPECT().Update(centerID, gomock.Any()).Return(dbErr)

	// Act
	err := s.service.Update(centerID, updates)

	// Assert
	s.Error(err)
	s.ErrorContains(err, "updating print center id 1")
}

// ============================================================================
// UpdateStatus Tests
// ============================================================================

func (s *PrintCenterServiceTestSuite) TestUpdateStatus_Success() {
	// Arrange
	var centerID uint = 1
	newStatus := entity.StatusApproved
	existingCenter := &entity.PrintCenter{ID: 1, Status: entity.StatusPending}

	s.mockRepo.EXPECT().FindByID(centerID).Return(existingCenter, nil)
	s.mockRepo.EXPECT().Update(centerID, gomock.Any()).DoAndReturn(func(id uint, updates map[string]any) error {
		s.Contains(updates, "status")
		s.Contains(updates, "updated_at")
		s.Equal(newStatus, updates["status"])
		return nil
	})

	// Act
	err := s.service.UpdateStatus(centerID, newStatus)

	// Assert
	s.NoError(err)
}

func (s *PrintCenterServiceTestSuite) TestUpdateStatus_NotFound() {
	// Arrange
	var centerID uint = 1
	s.mockRepo.EXPECT().FindByID(centerID).Return(nil, gorm.ErrRecordNotFound)

	// Act
	err := s.service.UpdateStatus(centerID, entity.StatusApproved)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrPrintCenterNotFound)
}

func (s *PrintCenterServiceTestSuite) TestUpdateStatus_DatabaseError() {
	// Arrange
	var centerID uint = 1
	existingCenter := &entity.PrintCenter{ID: 1}
	dbErr := errors.New("database error")

	s.mockRepo.EXPECT().FindByID(centerID).Return(existingCenter, nil)
	s.mockRepo.EXPECT().Update(centerID, gomock.Any()).Return(dbErr)

	// Act
	err := s.service.UpdateStatus(centerID, entity.StatusApproved)

	// Assert
	s.Error(err)
	s.Equal(dbErr, err)
}

// ============================================================================
// Delete Tests
// ============================================================================

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

func (s *PrintCenterServiceTestSuite) TestDelete_DatabaseError() {
	// Arrange
	var centerID uint = 1
	dbErr := errors.New("database error")
	s.mockRepo.EXPECT().Delete(centerID).Return(dbErr)

	// Act
	err := s.service.Delete(centerID)

	// Assert
	s.Error(err)
	s.ErrorContains(err, "failed to delete print center id 1")
}

// ============================================================================
// Helper Methods for Future Extensions
// ============================================================================

func (s *PrintCenterServiceTestSuite) createTestPrintCenter() *entity.PrintCenter {
	return &entity.PrintCenter{
		ID:   1,
		Name: "Test Print Center",
		Address: entity.Address{
			Number: "1",
			Type:   "Avenue",
			Street: "Kimba SABI N'GOYE",
			City:   "Kandi",
		},
		Status:    entity.StatusApproved,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *PrintCenterServiceTestSuite) createMultipleTestCenters(count int) []entity.PrintCenter {
	centers := make([]entity.PrintCenter, count)
	for i := range count {
		centers[i] = entity.PrintCenter{
			ID:   uint(i + 1),
			Name: fmt.Sprintf("Test Center %d", i+1),
			Address: entity.Address{ // Fix: strconv.FormatInt expects int64, not int
				Number: strconv.FormatInt(int64(i), 10),
				Type:   "Avenue",
				Street: "Kimba SABI N'GOYE",
				City:   "Kandi",
			},
			Status: entity.StatusApproved,
		}
	}
	return centers
}
