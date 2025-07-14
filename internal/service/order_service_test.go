package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/mocks"
	"github.com/kimbasn/printly/internal/service"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type OrderServiceTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	orderRepo       *mocks.MockOrderRepository
	printCenterRepo *mocks.MockPrintCenterRepository
	userRepo        *mocks.MockUserRepository
	service         service.OrderService
}

func (s *OrderServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.orderRepo = mocks.NewMockOrderRepository(s.ctrl)
	s.printCenterRepo = mocks.NewMockPrintCenterRepository(s.ctrl)
	s.userRepo = mocks.NewMockUserRepository(s.ctrl)
	s.service = service.NewOrderService(
		s.orderRepo,
		s.printCenterRepo,
		s.userRepo,
	)
}

func (s *OrderServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestOrderService(t *testing.T) {
	suite.Run(t, new(OrderServiceTestSuite))
}

// Test Cases for CreateOrder
func (s *OrderServiceTestSuite) TestCreateOrder_Success() {
	// Arrange
	userUID := "user-123"
	var centerID uint = 1
	req := dto.CreateOrderRequest{
		PrintMode:     entity.PrePrint,
		Documents: []dto.DocumentRequest{
			{
				FileName: "doc1.pdf",
				MimeType: "application/pdf",
				Size:     1024,
			},
		},
	}
	printCenter := &entity.PrintCenter{ID: centerID}

	s.printCenterRepo.EXPECT().FindByID(centerID).Return(printCenter, nil)
	s.orderRepo.EXPECT().FindByCode(gomock.Any()).Return(nil, nil) // For unique code generation
	s.orderRepo.EXPECT().Save(gomock.Any()).DoAndReturn(func(order *entity.Order) error {
		s.Equal(userUID, order.UserUID)
		s.Equal(centerID, order.PrintCenterID)
		s.Equal(entity.StatusAwaitingDocument, order.Status)
		s.Len(order.Documents, 1)
		s.Equal("doc1.pdf", order.Documents[0].FileName)
		return nil
	})

	// Act
	order, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.NoError(err)
	s.NotNil(order)
	s.NotEmpty(order.Code)
}

func (s *OrderServiceTestSuite) TestCreateOrder_PrintCenterNotFound() {
	// Arrange
	userUID := "user-123"
	var centerID uint = 1
	req := dto.CreateOrderRequest{}

	s.printCenterRepo.EXPECT().FindByID(centerID).Return(nil, nil)

	// Act
	_, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrPrintCenterNotFound)
}

func (s *OrderServiceTestSuite) TestCreateOrder_SaveError() {
	// Arrange
	userUID := "user-123"
	var centerID uint = 1
	req := dto.CreateOrderRequest{}
	printCenter := &entity.PrintCenter{ID: centerID}
	dbErr := errors.New("database save error")

	s.printCenterRepo.EXPECT().FindByID(centerID).Return(printCenter, nil)
	s.orderRepo.EXPECT().FindByCode(gomock.Any()).Return(nil, nil) // For unique code generation
	s.orderRepo.EXPECT().Save(gomock.Any()).Return(dbErr)

	// Act
	_, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.Error(err)
	s.ErrorContains(err, dbErr.Error())
}

func (s *OrderServiceTestSuite) TestCreateOrder_PrintCenterFindError() {
	// Arrange
	userUID := "user-123"
	var centerID uint = 1
	req := dto.CreateOrderRequest{}
	dbErr := errors.New("db find error")

	s.printCenterRepo.EXPECT().FindByID(centerID).Return(nil, dbErr)

	// Act
	_, err := s.service.CreateOrder(userUID, centerID, req)

	// Assert
	s.Error(err)
	s.ErrorContains(err, "failed to verify print center")
}

// Test Cases for GetOrderByID
func (s *OrderServiceTestSuite) TestGetOrderByID_Success() {
	// Arrange
	var orderID uint = 1
	expectedOrder := &entity.Order{ID: orderID}
	s.orderRepo.EXPECT().FindByID(orderID).Return(expectedOrder, nil)

	// Act
	order, err := s.service.GetOrderByID(orderID)

	// Assert
	s.NoError(err)
	s.Equal(expectedOrder, order)
}

func (s *OrderServiceTestSuite) TestGetOrderByID_NotFound() {
	// Arrange
	var orderID uint = 1
	s.orderRepo.EXPECT().FindByID(orderID).Return(nil, nil)

	// Act
	_, err := s.service.GetOrderByID(orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrOrderNotFound)
}

// Test Cases for GetOrderByCode
func (s *OrderServiceTestSuite) TestGetOrderByCode_Success() {
	// Arrange
	code := "ABC123"
	expectedOrder := &entity.Order{Code: code}
	s.orderRepo.EXPECT().FindByCode(code).Return(expectedOrder, nil)

	// Act
	order, err := s.service.GetOrderByCode(code)

	// Assert
	s.NoError(err)
	s.Equal(expectedOrder, order)
}

func (s *OrderServiceTestSuite) TestGetOrderByCode_NotFound() {
	// Arrange
	code := "ABC123"
	s.orderRepo.EXPECT().FindByCode(code).Return(nil, nil)

	// Act
	_, err := s.service.GetOrderByCode(code)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrOrderNotFound)
}

// Test Cases for UpdateOrderStatus
func (s *OrderServiceTestSuite) TestUpdateOrderStatus_Success() {
	// Arrange
	var orderID uint = 1
	newStatus := entity.StatusPaid
	existingOrder := &entity.Order{ID: orderID}

	s.orderRepo.EXPECT().FindByID(orderID).Return(existingOrder, nil)
	s.orderRepo.EXPECT().Update(orderID, gomock.Any()).DoAndReturn(func(_ uint, updates map[string]interface{}) error {
		s.Equal(newStatus, updates["status"])
		s.WithinDuration(time.Now(), updates["updated_at"].(time.Time), time.Second)
		return nil
	})

	// Act
	err := s.service.UpdateOrderStatus(orderID, newStatus)

	// Assert
	s.NoError(err)
}

func (s *OrderServiceTestSuite) TestUpdateOrderStatus_OrderNotFound() {
	// Arrange
	var orderID uint = 1
	newStatus := entity.StatusPaid

	s.orderRepo.EXPECT().FindByID(orderID).Return(nil, nil)

	// Act
	err := s.service.UpdateOrderStatus(orderID, newStatus)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrOrderNotFound)
}

func (s *OrderServiceTestSuite) TestUpdateOrderStatus_UpdateError() {
	// Arrange
	var orderID uint = 1
	newStatus := entity.StatusPaid
	dbErr := errors.New("db update error")

	s.orderRepo.EXPECT().FindByID(orderID).Return(&entity.Order{ID: orderID}, nil)
	s.orderRepo.EXPECT().Update(orderID, gomock.Any()).Return(dbErr)

	// Act
	err := s.service.UpdateOrderStatus(orderID, newStatus)

	// Assert
	s.Error(err)
	s.ErrorContains(err, dbErr.Error())
}

// Test Cases for DeleteOrder
func (s *OrderServiceTestSuite) TestDeleteOrder_Success() {
	// Arrange
	var orderID uint = 1
	s.orderRepo.EXPECT().Delete(orderID).Return(nil)

	// Act
	err := s.service.DeleteOrder(orderID)

	// Assert
	s.NoError(err)
}

func (s *OrderServiceTestSuite) TestDeleteOrder_NotFound() {
	// Arrange
	var orderID uint = 1
	s.orderRepo.EXPECT().Delete(orderID).Return(gorm.ErrRecordNotFound)

	// Act
	err := s.service.DeleteOrder(orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, ierrors.ErrOrderNotFound)
}
