package store

import (
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// OrderStore handles database operations for Orders.
type OrderStore struct {
	db *gorm.DB
}

// NewOrderStore creates a new OrderStore.
func NewOrderStore(db *gorm.DB) *OrderStore {
	return &OrderStore{db: db}
}

// CreateOrder creates a new order and its associated details.
// This should be done within a transaction to ensure data integrity.
func (s *OrderStore) CreateOrder(order *core.Order) error {
	return s.db.Create(order).Error
}

// GetOrderByID retrieves an order by its ID, preloading all related data.
func (s *OrderStore) GetOrderByID(id uuid.UUID) (*core.Order, error) {
	var order core.Order
	err := s.db.
		Preload("OrderDetails.Variant").
		Preload("Payments").
		Preload("Delivery").
		Preload("Voucher").
		First(&order, "id = ?", id).Error
	return &order, err
}

// UpdateOrder updates an existing order.
func (s *OrderStore) UpdateOrder(order *core.Order) error {
	return s.db.Save(order).Error
}
