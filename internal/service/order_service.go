package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"time"

	"go.uber.org/zap"

	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/entity"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/repository"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../mocks/mock_order_service.go -package=mocks github.com/kimbasn/printly/internal/service OrderService

// OrderService defines the interface for order-related business logic.
type OrderService interface {
	CreateOrder(userUID string, centerID uint, req dto.CreateOrderRequest) (*entity.Order, error)
	GetOrderByID(id uint) (*entity.Order, error)
	GetOrderByCode(code string) (*entity.Order, error)
	GetOrdersForCenter(centerID uint) ([]entity.Order, error)
	GetOrdersForUser(userUID string) ([]entity.Order, error)
	GetAllOrders() ([]entity.Order, error)
	UpdateOrderStatus(orderID uint, status entity.OrderStatus, updatedBy string) error
	CancelOrder(orderID uint, userUID string) error
	DeleteOrder(orderID uint) error
	CalculateOrderCost(orderID uint) (int64, error)
}

type orderService struct {
	orderRepo       repository.OrderRepository
	printCenterRepo repository.PrintCenterRepository
	userRepo        repository.UserRepository
	logger          *zap.Logger
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(orderRepo repository.OrderRepository, printCenterRepo repository.PrintCenterRepository, userRepo repository.UserRepository, logger *zap.Logger) OrderService {
	return &orderService{
		orderRepo:       orderRepo,
		printCenterRepo: printCenterRepo,
		userRepo:        userRepo,
		logger:          logger,
	}
}

// CreateOrder handles the business logic for creating a new order.
func (s *orderService) CreateOrder(userUID string, centerID uint, req dto.CreateOrderRequest) (*entity.Order, error) {
	s.logger.Info("Creating order", zap.String("userUID", userUID), zap.Uint("centerID", centerID))

	// 1. Verify print center exists and is operational
	center, err := s.printCenterRepo.FindByID(centerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ierrors.ErrPrintCenterNotFound
		}
		return nil, fmt.Errorf("failed to verify print center: %w", err)
	}
	if center.Status != entity.StatusApproved {
		return nil, ierrors.ErrPrintCenterNotOperational
	}

	// 2. Generate a unique pickup code
	code, err := s.generateUniquePickupCode(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pickup code: %w", err)
	}

	// 3. Create and save the order
	order := &entity.Order{
		UserUID:       userUID,
		PrintCenterID: centerID,
		Status:        entity.StatusPendingPayment,
		Code:          code,
		CreatedBy:     userUID,
		UpdatedBy:     userUID,
		Documents:     make([]entity.Document, len(req.Documents)),
	}

	if err := s.orderRepo.Save(order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	s.logger.Info("Order created successfully", zap.Uint("orderID", order.ID), zap.String("code", order.Code))

	return order, nil
}

// GetOrderByID retrieves an order by its ID.
func (s *orderService) GetOrderByID(id uint) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ierrors.ErrOrderNotFound
		}
		return nil, fmt.Errorf("getting order by id %d: %w", id, err)
	}
	return order, nil
}

// GetOrderByCode retrieves an order by its pickup code.
func (s *orderService) GetOrderByCode(code string) (*entity.Order, error) {
	order, err := s.orderRepo.FindByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ierrors.ErrOrderNotFound
		}
		return nil, fmt.Errorf("getting order by code %s: %w", code, err)
	}
	return order, nil
}

// GetOrdersForCenter retrieves all orders for a specific print center.
func (s *orderService) GetOrdersForCenter(centerID uint) ([]entity.Order, error) {
	orders, err := s.orderRepo.FindByCenterID(centerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders for center %d: %w", centerID, err)
	}
	return orders, nil
}

// GetOrdersForUser retrieves all orders for a specific user.
func (s *orderService) GetOrdersForUser(userUID string) ([]entity.Order, error) {
	orders, err := s.orderRepo.FindByUserUID(userUID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders for user %s: %w", userUID, err)
	}
	return orders, nil
}

// GetAllOrders retrieves all orders (for admin use).
func (s *orderService) GetAllOrders() ([]entity.Order, error) {
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all orders: %w", err)
	}
	return orders, nil
}

// UpdateOrderStatus updates the status of a specific order.
func (s *orderService) UpdateOrderStatus(orderID uint, status entity.OrderStatus, updatedBy string) error {
	s.logger.Info("Updating order status",
		zap.Uint("orderID", orderID),
		zap.String("status", string(status)),
		zap.String("updatedBy", updatedBy))

	if _, err := s.GetOrderByID(orderID); err != nil {
		return err // Return ErrOrderNotFound if it doesn't exist
	}

	updates := map[string]any{
		"status":     status,
		"updated_by": updatedBy,
		"updated_at": time.Now(),
	}

	err := s.orderRepo.Update(orderID, updates)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("Order status updated successfully", zap.Uint("orderID", orderID), zap.String("status", string(status)))
	return nil
}

// CancelOrder cancels an order if it's in a cancellable state.
func (s *orderService) CancelOrder(orderID uint, userUID string) error {
	s.logger.Info("Canceling order", zap.Uint("orderID", orderID), zap.String("userUID", userUID))

	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return err // Return ErrOrderNotFound if it doesn't exist
	}

	// Verify that the user owns this order
	// TODO: Need to handle admin cancelling
	if order.UserUID != userUID {
		return ierrors.ErrUnauthorized
	}

	// Check if the order can be cancelled
	if !order.CanCancel() {
		return ierrors.ErrOrderCannotBeCancelled
	}

	err = s.UpdateOrderStatus(orderID, entity.StatusCancelled, userUID)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	s.logger.Info("Order cancelled successfully", zap.Uint("orderID", orderID))
	return nil
}

// DeleteOrder removes an order from the database.
func (s *orderService) DeleteOrder(orderID uint) error {
	s.logger.Info("Deleting order", zap.Uint("orderID", orderID))

	err := s.orderRepo.Delete(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ierrors.ErrOrderNotFound
		}
		return fmt.Errorf("failed to delete order id %d: %w", orderID, err)
	}

	s.logger.Info("Order deleted successfully", zap.Uint("orderID", orderID))
	return nil
}

// CalculateOrderCost calculates the total cost of an order based on its documents and print options.
func (s *orderService) CalculateOrderCost(orderID uint) (int64, error) {
	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return 0, err
	}

	var totalCost int64

	for _, doc := range order.Documents {
		// Base cost calculation (example: $0.10 per page)
		baseCostPerPage := int64(10) // 10 cents in the smallest currency unit

		// Calculate pages based on document size (this is a simplified example)
		// In a real implementation, you would need to determine the actual page count
		estimatedPages := int64(1) // Default to 1 page
		if doc.Size > 0 {
			// Rough estimation: 50KB per page (adjust based on your requirements)
			estimatedPages = (doc.Size + 50000 - 1) / 50000
			if estimatedPages == 0 {
				estimatedPages = 1
			}
		}

		docCost := baseCostPerPage * estimatedPages

		// Apply print options modifiers
		if doc.PrintOptions.Color == entity.Color {
			docCost *= 3 // Color printing costs 3x more
		}
		if doc.PrintOptions.DoubleSided {
			docCost = docCost * 6 / 10 // 40% discount for double-sided
		}
		if doc.PrintOptions.Copies > 1 {
			docCost *= int64(doc.PrintOptions.Copies)
		}

		totalCost += docCost
	}

	s.logger.Info("Order cost calculated", zap.Uint("orderID", orderID), zap.Int64("totalCost", totalCost))
	return totalCost, nil
}

// generateUniquePickupCode creates a random alphanumeric string of a given length.
func (s *orderService) generateUniquePickupCode(length int) (string, error) {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const maxRetries = 10 // To prevent infinite loop

	for range maxRetries {
		b := make([]byte, length)
		if _, err := io.ReadFull(rand.Reader, b); err != nil {
			return "", fmt.Errorf("failed to read random bytes for code: %w", err)
		}
		for i := range length {
			b[i] = table[int(b[i])%len(table)]
		}
		code := string(b)

		// Check for uniqueness
		existing, err := s.orderRepo.FindByCode(code)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return code, nil // Code is unique
			}
			return "", fmt.Errorf("failed to check for code uniqueness: %w", err)
		}
		if existing == nil {
			return code, nil // Code is unique
		}
	}

	return "", errors.New("failed to generate a unique pickup code after multiple retries")
}
