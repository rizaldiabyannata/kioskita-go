package store

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

type OrderStore struct {
	db *gorm.DB
}

func NewOrderStore(db *gorm.DB) *OrderStore {
	return &OrderStore{db: db}
}

// GetCartByUserID finds an active cart for a user.
func (s *OrderStore) GetCartByUserID(userID uuid.UUID) (*core.Order, error) {
	var order core.Order
	err := s.db.Where("user_id = ? AND status = ?", userID, core.StatusCart).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// CreateOrder creates a new order (cart).
func (s *OrderStore) CreateOrder(order *core.Order) error {
	return s.db.Create(order).Error
}

// AddOrderItem adds a product to an order.
func (s *OrderStore) AddOrderItem(item *core.OrderItem) error {
	return s.db.Create(item).Error
}

// GetOrderItem finds a specific item in an order.
func (s *OrderStore) GetOrderItem(orderID, productID uuid.UUID) (*core.OrderItem, error) {
	var item core.OrderItem
	err := s.db.Where("order_id = ? AND product_id = ?", orderID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateOrderItem updates an item's quantity.
func (s *OrderStore) UpdateOrderItem(item *core.OrderItem) error {
	return s.db.Model(item).Update("quantity", item.Quantity).Error
}

// DeleteOrderItem removes an item from an order.
func (s *OrderStore) DeleteOrderItem(itemID uuid.UUID) error {
	result := s.db.Delete(&core.OrderItem{}, "id = ?", itemID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetCartViewByUserID retrieves the full cart view with products.
func (s *OrderStore) GetCartViewByUserID(userID uuid.UUID) (*core.CartView, error) {
	var order core.Order
	// Use Preload to solve the N+1 query problem
	err := s.db.Preload("Items.Product").Where("user_id = ? AND status = ?", userID, core.StatusCart).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No cart found is not an application error
		}
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	cartView := &core.CartView{
		Order: order,
		Items: []core.CartItemView{},
	}

	var totalAmount int64
	for _, item := range order.Items {
		// Ensure product is not nil, in case of data inconsistency
		if item.Product.ID != uuid.Nil {
			totalAmount += item.Product.Price * int64(item.Quantity)
			cartView.Items = append(cartView.Items, core.CartItemView{
				OrderItem: item,
				Product:   item.Product,
			})
		}
	}

	// Update the total amount of the order if it's different
	if order.TotalAmount != totalAmount {
		err := s.db.Model(&order).Update("total_amount", totalAmount).Error
		if err != nil {
			// Log or handle the error, but we can still return the view
			fmt.Printf("Warning: failed to update total amount for order %s: %v\n", order.ID, err)
		}
		cartView.Order.TotalAmount = totalAmount
	}

	return cartView, nil
}

// GetOrderItemByID retrieves a single order item by its ID.
func (s *OrderStore) GetOrderItemByID(itemID uuid.UUID) (*core.OrderItem, error) {
	var item core.OrderItem
	err := s.db.First(&item, "id = ?", itemID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
