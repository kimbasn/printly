package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"

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
	GetAllOrders() ([]entity.Order, error)
	UpdateOrderStatus(orderID uint, status entity.OrderStatus) error
	DeleteOrder(orderID uint) error
}

type orderService struct {
	orderRepo       repository.OrderRepository
	printCenterRepo repository.PrintCenterRepository
	userRepo        repository.UserRepository
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(orderRepo repository.OrderRepository, printCenterRepo repository.PrintCenterRepository, userRepo repository.UserRepository) OrderService {
	return &orderService{
		orderRepo:       orderRepo,
		printCenterRepo: printCenterRepo,
		userRepo:        userRepo,
	}
}

// CreateOrder handles the business logic for creating a new order.
func (s *orderService) CreateOrder(userUID string, centerID uint, req dto.CreateOrderRequest) (*entity.Order, error) {
	center, err := s.printCenterRepo.FindByID(centerID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify print center: %w", err)
	}
	if center == nil {
		return nil, ierrors.ErrPrintCenterNotFound
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
		Status:        entity.StatusAwaitingDocument, // Initial status from design doc
		PrintMode:     req.PrintMode,
		Code:          code,
		Documents:     make([]entity.Document, len(req.Documents)),
	}

	for i, docReq := range req.Documents {
		order.Documents[i] = entity.Document{
			FileName:     docReq.FileName,
			MimeType:     docReq.MimeType,
			Size:         docReq.Size,
			PrintOptions: docReq.PrintOptions,
		}
	}

	if err := s.orderRepo.Save(order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return order, nil
}

// GetOrderByID retrieves an order by its ID.
func (s *orderService) GetOrderByID(id uint) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("getting order by id %d: %w", id, err)
	}
	if order == nil {
		return nil, ierrors.ErrOrderNotFound
	}
	return order, nil
}

// GetOrderByCode retrieves an order by its pickup code.
func (s *orderService) GetOrderByCode(code string) (*entity.Order, error) {
	order, err := s.orderRepo.FindByCode(code)
	if err != nil {
		return nil, fmt.Errorf("getting order by code %s: %w", code, err)
	}
	if order == nil {
		return nil, ierrors.ErrOrderNotFound
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

// GetAllOrders retrieves all orders (for admin use).
func (s *orderService) GetAllOrders() ([]entity.Order, error) {
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all orders: %w", err)
	}
	return orders, nil
}

// UpdateOrderStatus updates the status of a specific order.
func (s *orderService) UpdateOrderStatus(orderID uint, status entity.OrderStatus) error {
	if _, err := s.GetOrderByID(orderID); err != nil {
		return err // Return ErrOrderNotFound if it doesn't exist
	}

	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
	}

	return s.orderRepo.Update(orderID, updates)
}

// DeleteOrder removes an order from the database.
func (s *orderService) DeleteOrder(orderID uint) error {
	err := s.orderRepo.Delete(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ierrors.ErrOrderNotFound
		}
		return fmt.Errorf("failed to delete order id %d: %w", orderID, err)
	}
	return nil
}

// generateUniquePickupCode creates a random alphanumeric string of a given length.
func (s *orderService) generateUniquePickupCode(length int) (string, error) {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const maxRetries = 10 // TO prevent infinit loop

	for range maxRetries {
		b := make([]byte, length)
		if _, err := io.ReadFull(rand.Reader, b); err != nil {
			return "", fmt.Errorf("failed to read random bytes for code %w", err)
		}
		for i := range length {
			b[i] = table[int(b[i])%len(table)]
		}
		code := string(b)

		// check for uniqueness
		existing, err := s.orderRepo.FindByCode(code)
		if err != nil {
			return "", fmt.Errorf("failed to check for code uniqueness: %w", err)
		}
		if existing == nil {
			return code, nil // Code is unique
		}
	}

	return "", errors.New("failed to generate a uniquepickup code after multiple retries")
}
