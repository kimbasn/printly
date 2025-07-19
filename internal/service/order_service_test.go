package service_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/mocks"
	"github.com/kimbasn/printly/internal/service"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderServiceTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	orderRepo       *mocks.MockOrderRepository
	printCenterRepo *mocks.MockPrintCenterRepository
	userRepo        *mocks.MockUserRepository
	service         service.OrderService
	logger          *zap.Logger
}

func (s *OrderServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.orderRepo = mocks.NewMockOrderRepository(s.ctrl)
	s.printCenterRepo = mocks.NewMockPrintCenterRepository(s.ctrl)
	s.userRepo = mocks.NewMockUserRepository(s.ctrl)
	s.logger = zap.NewNop()

	s.service = service.NewOrderService(
		s.orderRepo,
		s.printCenterRepo,
		s.userRepo,
		s.logger,
	)
}

func (s *OrderServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestOrderService(t *testing.T) {
	suite.Run(t, new(OrderServiceTestSuite))
}

// ============================================================================
// CreateOrder Tests
// ============================================================================

func (s *OrderServiceTestSuite) TestCreateOrder_Success() {
	// Arrange
	userUID := "test-user-123"
	centerID := uint(1)
	req := dto.CreateOrderRequest{
		Documents: []dto.CreateDocumentRequest{
			{
				FileName: "test.pdf",
				Size:     1024,
				MimeType: "application/pdf",
				PrintOptions: entity.PrintOptions{
					Color:       entity.BlackAndWhite,
					DoubleSided: false,
					Copies:      1,
				},
			},
		},
	}

	center := &entity.PrintCenter{
		ID:     centerID,
		Status: entity.StatusApproved,
	}

	// Mock expectations
	s.printCenterRepo.EXPECT().
		FindByID(centerID).
		Return(center, nil)

	s.orderRepo.EXPECT().
		FindByCode(gomock.Any()).
		Return(nil, gorm.ErrRecordNotFound) // Code is unique

	s.orderRepo.EXPECT().
		Save(gomock.Any()).
		DoAndReturn(func(order *entity.Order) error {
			order.ID = 1 // Simulate database ID assignment
			return nil
		})

	// Act
	result, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(userUID, result.UserUID)
	s.Equal(centerID, result.PrintCenterID)
	s.Equal(entity.StatusPendingPayment, result.Status)
	s.NotEmpty(result.Code)
	s.Len(result.Code, 6)
}

func (s *OrderServiceTestSuite) TestCreateOrder_PrintCenterNotFound() {
	// Arrange
	userUID := "test-user-123"
	centerID := uint(999)
	req := dto.CreateOrderRequest{Documents: []dto.CreateDocumentRequest{}}

	// Mock expectations
	s.printCenterRepo.EXPECT().
		FindByID(centerID).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	result, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Equal(ierrors.ErrPrintCenterNotFound, err)
}

func (s *OrderServiceTestSuite) TestCreateOrder_PrintCenterNotOperational() {
	// Arrange
	userUID := "test-user-123"
	centerID := uint(1)
	req := dto.CreateOrderRequest{Documents: []dto.CreateDocumentRequest{}}

	center := &entity.PrintCenter{
		ID:     centerID,
		Status: entity.StatusPending, // Not approved
	}

	// Mock expectations
	s.printCenterRepo.EXPECT().
		FindByID(centerID).
		Return(center, nil)

	// Act
	result, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Equal(ierrors.ErrPrintCenterNotOperational, err)
}

func (s *OrderServiceTestSuite) TestCreateOrder_SaveOrderError() {
	// Arrange
	userUID := "test-user-123"
	centerID := uint(1)
	req := dto.CreateOrderRequest{Documents: []dto.CreateDocumentRequest{}}

	center := &entity.PrintCenter{
		ID:     centerID,
		Status: entity.StatusApproved,
	}

	// Mock expectations
	s.printCenterRepo.EXPECT().
		FindByID(centerID).
		Return(center, nil)

	s.orderRepo.EXPECT().
		FindByCode(gomock.Any()).
		Return(nil, gorm.ErrRecordNotFound)

	s.orderRepo.EXPECT().
		Save(gomock.Any()).
		Return(errors.New("database error"))

	// Act
	result, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to save order")
}

// ============================================================================
// GetOrderByID Tests
// ============================================================================

func (s *OrderServiceTestSuite) TestGetOrderByID_Success() {
	// Arrange
	orderID := uint(1)
	expectedOrder := &entity.Order{
		ID:      orderID,
		UserUID: "test-user-123",
		Status:  entity.StatusPendingPayment,
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(expectedOrder, nil)

	// Act
	result, err := s.service.GetOrderByID(orderID)

	// Assert
	s.NoError(err)
	s.Equal(expectedOrder, result)
}

func (s *OrderServiceTestSuite) TestGetOrderByID_NotFound() {
	// Arrange
	orderID := uint(999)

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	result, err := s.service.GetOrderByID(orderID)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Equal(ierrors.ErrOrderNotFound, err)
}

// ============================================================================
// GetOrderByCode Tests
// ============================================================================
func (s *OrderServiceTestSuite) TestGetOrderByCode_Success() {
	// Arrange
	code := "ABC123"
	expectedOrder := &entity.Order{
		ID:   1,
		Code: code,
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByCode(code).
		Return(expectedOrder, nil)

	// Act
	result, err := s.service.GetOrderByCode(code)

	// Assert
	s.NoError(err)
	s.Equal(expectedOrder, result)
}

func (s *OrderServiceTestSuite) TestGetOrderByCode_NotFound() {
	// Arrange
	code := "INVALID"

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByCode(code).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	result, err := s.service.GetOrderByCode(code)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Equal(ierrors.ErrOrderNotFound, err)
}

// ============================================================================
// GetOrdersForCenter Tests
// ============================================================================
func (s *OrderServiceTestSuite) TestGetOrdersForCenter_Success() {
	// Arrange
	centerID := uint(1)
	expectedOrders := []entity.Order{
		{ID: 1, PrintCenterID: centerID},
		{ID: 2, PrintCenterID: centerID},
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByCenterID(centerID).
		Return(expectedOrders, nil)

	// Act
	result, err := s.service.GetOrdersForCenter(centerID)

	// Assert
	s.NoError(err)
	s.Equal(expectedOrders, result)
}

func (s *OrderServiceTestSuite) TestGetOrdersForCenter_Error() {
	// Arrange
	centerID := uint(1)

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByCenterID(centerID).
		Return(nil, errors.New("database error"))

	// Act
	result, err := s.service.GetOrdersForCenter(centerID)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to fetch orders for center")
}

// ============================================================================
// GetOrdersForUser Tests
// ============================================================================
func (s *OrderServiceTestSuite) TestGetOrdersForUser_Success() {
	// Arrange
	userUID := "test-user-123"
	expectedOrders := []entity.Order{
		{ID: 1, UserUID: userUID},
		{ID: 2, UserUID: userUID},
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByUserUID(userUID).
		Return(expectedOrders, nil)

	// Act
	result, err := s.service.GetOrdersForUser(userUID)

	// Assert
	s.NoError(err)
	s.Equal(expectedOrders, result)
}

// ============================================================================
// GetAllOrders Tests
// ============================================================================
func (s *OrderServiceTestSuite) TestGetAllOrders_Success() {
	// Arrange
	expectedOrders := []entity.Order{
		{ID: 1, UserUID: "user1"},
		{ID: 2, UserUID: "user2"},
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindAll().
		Return(expectedOrders, nil)

	// Act
	result, err := s.service.GetAllOrders()

	// Assert
	s.NoError(err)
	s.Equal(expectedOrders, result)
}

// ============================================================================
// UpdateOrderStatus Tests
// ============================================================================
func (s *OrderServiceTestSuite) TestUpdateOrderStatus_Success() {
	// Arrange
	orderID := uint(1)
	status := entity.StatusCompleted
	updatedBy := "admin-123"

	existingOrder := &entity.Order{
		ID:      orderID,
		Status:  entity.StatusPendingPayment,
		UserUID: "test-user-123",
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(existingOrder, nil)

	s.orderRepo.EXPECT().
		Update(orderID, gomock.Any()).
		DoAndReturn(func(id uint, updates map[string]any) error {
			// Verify the updates contain the expected fields
			s.Equal(status, updates["status"])
			s.Equal(updatedBy, updates["updated_by"])
			s.NotNil(updates["updated_at"])
			return nil
		})

	// Act
	err := s.service.UpdateOrderStatus(orderID, status, updatedBy)

	// Assert
	s.NoError(err)
}

func (s *OrderServiceTestSuite) TestUpdateOrderStatus_OrderNotFound() {
	// Arrange
	orderID := uint(999)
	status := entity.StatusCompleted
	updatedBy := "admin-123"

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	err := s.service.UpdateOrderStatus(orderID, status, updatedBy)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrOrderNotFound, err)
}

func (s *OrderServiceTestSuite) TestUpdateOrderStatus_UpdateError() {
	// Arrange
	var orderID uint = 1
	userUID := "test-user-123"
	newStatus := entity.StatusPaid
	dbErr := errors.New("db update error")

	s.orderRepo.EXPECT().FindByID(orderID).Return(&entity.Order{ID: orderID}, nil)
	s.orderRepo.EXPECT().Update(orderID, gomock.Any()).Return(dbErr)

	// Act
	err := s.service.UpdateOrderStatus(orderID, newStatus, userUID)

	// Assert
	s.Error(err)
	s.ErrorContains(err, dbErr.Error())
}

// ============================================================================
// CancelOrder Tests
// ============================================================================

func (s *OrderServiceTestSuite) TestCancelOrder_Success() {
	// Arrange
	orderID := uint(1)
	userUID := "test-user-123"

	existingOrder := &entity.Order{
		ID:      orderID,
		UserUID: userUID,
		Status:  entity.StatusPendingPayment,
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(existingOrder, nil).
		Times(2)

	s.orderRepo.EXPECT().
		Update(orderID, gomock.Any()).
		Return(nil)

	// Act
	err := s.service.CancelOrder(orderID, userUID)

	// Assert
	s.NoError(err)
}

func (s *OrderServiceTestSuite) TestCancelOrder_Unauthorized() {
	// Arrange
	orderID := uint(1)
	userUID := "wrong-user-123"

	existingOrder := &entity.Order{
		ID:      orderID,
		UserUID: "test-user-123", // Different user
		Status:  entity.StatusPendingPayment,
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(existingOrder, nil)

	// Act
	err := s.service.CancelOrder(orderID, userUID)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrUnauthorized, err)
}

func (s *OrderServiceTestSuite) TestCancelOrder_CannotBeCancelled() {
	// Arrange
	orderID := uint(1)
	userUID := "test-user-123"

	existingOrder := &entity.Order{
		ID:      orderID,
		UserUID: userUID,
		Status:  entity.StatusCompleted, // Cannot cancel completed orders
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(existingOrder, nil)

	// Act
	err := s.service.CancelOrder(orderID, userUID)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrOrderCannotBeCancelled, err)
}

// ============================================================================
// DeleteOrder Tests
// ============================================================================

func (s *OrderServiceTestSuite) TestDeleteOrder_Success() {
	// Arrange
	orderID := uint(1)

	// Mock expectations
	s.orderRepo.EXPECT().
		Delete(orderID).
		Return(nil)

	// Act
	err := s.service.DeleteOrder(orderID)

	// Assert
	s.NoError(err)
}

func (s *OrderServiceTestSuite) TestDeleteOrder_NotFound() {
	// Arrange
	orderID := uint(999)

	// Mock expectations
	s.orderRepo.EXPECT().
		Delete(orderID).
		Return(gorm.ErrRecordNotFound)

	// Act
	err := s.service.DeleteOrder(orderID)

	// Assert
	s.Error(err)
	s.Equal(ierrors.ErrOrderNotFound, err)
}

// ============================================================================
// CalculateOrderCost Tests
// ============================================================================

func (s *OrderServiceTestSuite) TestCalculateOrderCost_Success() {
	// Arrange
	orderID := uint(1)
	order := &entity.Order{
		ID: orderID,
		Documents: []entity.Document{
			{
				Size: 100000, // 100KB
				PrintOptions: entity.PrintOptions{
					Color:       entity.BlackAndWhite,
					DoubleSided: false,
					Copies:      1,
				},
			},
			{
				Size: 200000, // 200KB
				PrintOptions: entity.PrintOptions{
					Color:       entity.Color,
					DoubleSided: true,
					Copies:      2,
				},
			},
		},
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(order, nil)

	// Act
	cost, err := s.service.CalculateOrderCost(orderID)

	// Assert
	s.NoError(err)
	s.Greater(cost, int64(0))
	
	// Verify cost calculation logic
	// First document: 2 pages * 10 cents = 20 cents
	// Second document: 4 pages * 10 cents * 3 (color) * 0.6 (double-sided) * 2 (copies) = 144 cents
	expectedCost := int64(20 + 144)
	s.Equal(expectedCost, cost)
}

func (s *OrderServiceTestSuite) TestCalculateOrderCost_OrderNotFound() {
	// Arrange
	orderID := uint(999)

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(nil, gorm.ErrRecordNotFound)

	// Act
	cost, err := s.service.CalculateOrderCost(orderID)

	// Assert
	s.Error(err)
	s.Equal(int64(0), cost)
	s.Equal(ierrors.ErrOrderNotFound, err)
}

func (s *OrderServiceTestSuite) TestCalculateOrderCost_EmptyDocuments() {
	// Arrange
	orderID := uint(1)
	order := &entity.Order{
		ID:        orderID,
		Documents: []entity.Document{}, // No documents
	}

	// Mock expectations
	s.orderRepo.EXPECT().
		FindByID(orderID).
		Return(order, nil)

	// Act
	cost, err := s.service.CalculateOrderCost(orderID)

	// Assert
	s.NoError(err)
	s.Equal(int64(0), cost)
}

// ============================================================================
// generateUniquePickupCode Tests
// ============================================================================
func (s *OrderServiceTestSuite) TestCreateOrder_GenerateUniqueCode_RetryLogic() {
	// Arrange
	userUID := "test-user-123"
	centerID := uint(1)
	req := dto.CreateOrderRequest{Documents: []dto.CreateDocumentRequest{}}

	center := &entity.PrintCenter{
		ID:     centerID,
		Status: entity.StatusApproved,
	}

	existingOrder := &entity.Order{ID: 1, Code: "EXISTING"}

	// Mock expectations
	s.printCenterRepo.EXPECT().
		FindByID(centerID).
		Return(center, nil)

	// First call returns existing order (code collision)
	s.orderRepo.EXPECT().
		FindByCode(gomock.Any()).
		Return(existingOrder, nil)

	// Second call returns not found (unique code)
	s.orderRepo.EXPECT().
		FindByCode(gomock.Any()).
		Return(nil, gorm.ErrRecordNotFound)

	s.orderRepo.EXPECT().
		Save(gomock.Any()).
		DoAndReturn(func(order *entity.Order) error {
			order.ID = 1
			return nil
		})

	// Act
	result, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.NotEmpty(result.Code)
}