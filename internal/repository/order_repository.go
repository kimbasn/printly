package repository

import (
	"errors"
	"fmt"

	"github.com/kimbasn/printly/internal/entity"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../mocks/mock_order_repository.go -package=mocks github.com/kimbasn/printly/internal/repository OrderRepository

// OrderRepository defines the interface for order-related database operations.
type OrderRepository interface {
	Save(order *entity.Order) error
	FindByID(id uint) (*entity.Order, error)
	FindByCode(code string) (*entity.Order, error)
	FindByCenterID(centerID uint) ([]entity.Order, error)
	FindByUserUID(userUID string) ([]entity.Order, error)
	FindByStatus(status entity.OrderStatus) ([]entity.Order, error)
	FindAll() ([]entity.Order, error)
	Update(id uint, updates map[string]any) error
	Delete(id uint) error
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new instance of an OrderRepository.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Save creates a new order record in the database.
func (r *orderRepository) Save(order *entity.Order) error {
	if err := r.db.Create(order).Error; err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	return nil
}

// FindByID retrieves an order from the database by its primary key.
func (r *orderRepository) FindByID(id uint) (*entity.Order, error) {
	var order entity.Order
	result := r.db.First(&order, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch order with id %d: %w", id, result.Error)
	}
	return &order, nil
}

// FindByCode retrieves an order by its unique pickup code.
func (r *orderRepository) FindByCode(code string) (*entity.Order, error) {
	var order entity.Order
	result := r.db.First(&order, "code = ?", code)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch order with code %s: %w", code, result.Error)
	}
	return &order, nil
}

// FindByCenterID retrieves all orders associated with a specific print center.
func (r *orderRepository) FindByCenterID(centerID uint) ([]entity.Order, error) {
	var orders []entity.Order
	result := r.db.Find(&orders, "print_center_id = ?", centerID)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch orders for center id %d: %w", centerID, result.Error)
	}
	return orders, nil
}

func (r *orderRepository) FindByUserUID(userUID string) ([]entity.Order, error) {
	var orders []entity.Order
	result := r.db.Find(&orders, "user_uid = ?", userUID)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch orders for user uid %s: %w", userUID, result.Error)
	}
	return orders, nil
}

// FindByStatus retrieves all orders with a specific status.
func (r *orderRepository) FindByStatus(status entity.OrderStatus) ([]entity.Order, error) {
	var orders []entity.Order
	result := r.db.Find(&orders, "status = ?", status)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch orders with status %s: %w", status, result.Error)
	}
	return orders, nil
}

// FindAll retrieves all order records.
func (r *orderRepository) FindAll() ([]entity.Order, error) {
	var orders []entity.Order
	if err := r.db.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch all orders: %w", err)
	}
	return orders, nil
}

// Update modifies an existing order's record.
func (r *orderRepository) Update(id uint, updates map[string]any) error {
	result := r.db.Model(&entity.Order{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update order id %d: %w", id, result.Error)
	}
	return nil
}

// Delete removes an order from the database.
func (r *orderRepository) Delete(id uint) error {
	result := r.db.Delete(&entity.Order{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete order id %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
