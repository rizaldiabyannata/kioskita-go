package store

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
)

type OrderStore struct {
	db *sqlx.DB
}

func NewOrderStore(db *sqlx.DB) *OrderStore {
	return &OrderStore{db: db}
}

// GetCartByUserID finds an active cart for a user.
func (s *OrderStore) GetCartByUserID(userID uuid.UUID) (*core.Order, error) {
	var order core.Order
	query := `SELECT * FROM orders WHERE user_id = $1 AND status = $2`
	err := s.db.Get(&order, query, userID, core.StatusCart)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// CreateOrder creates a new order (cart).
func (s *OrderStore) CreateOrder(order *core.Order) error {
	query := `INSERT INTO orders (id, user_id, total_amount, status, created_at)
              VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(query, order.ID, order.UserID, order.TotalAmount, order.Status, order.CreatedAt)
	return err
}

// AddOrderItem adds a product to an order.
func (s *OrderStore) AddOrderItem(item *core.OrderItem) error {
	query := `INSERT INTO order_items (id, order_id, product_id, quantity, price_at_purchase, created_at)
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.Exec(query, item.ID, item.OrderID, item.ProductID, item.Quantity, item.PriceAtPurchase, item.CreatedAt)
	return err
}

// GetOrderItem finds a specific item in an order.
func (s *OrderStore) GetOrderItem(orderID, productID uuid.UUID) (*core.OrderItem, error) {
	var item core.OrderItem
	query := `SELECT * FROM order_items WHERE order_id = $1 AND product_id = $2`
	err := s.db.Get(&item, query, orderID, productID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateOrderItem updates an item's quantity.
func (s *OrderStore) UpdateOrderItem(item *core.OrderItem) error {
	query := `UPDATE order_items SET quantity = $1 WHERE id = $2`
	_, err := s.db.Exec(query, item.Quantity, item.ID)
	return err
}

// DeleteOrderItem removes an item from an order.
func (s *OrderStore) DeleteOrderItem(itemID uuid.UUID) error {
	query := `DELETE FROM order_items WHERE id = $1`
	res, err := s.db.Exec(query, itemID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetCartViewByUserID retrieves the full cart view with products.
func (s *OrderStore) GetCartViewByUserID(userID uuid.UUID) (*core.CartView, error) {
	// 1. Get the active cart order
	order, err := s.GetCartByUserID(userID)
	if err != nil {
		// If no cart exists, return a specific error or nil
		if err == sql.ErrNoRows {
			return nil, nil // Or a custom "cart not found" error
		}
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	// 2. Get all items for that order
	var items []core.OrderItem
	queryItems := `SELECT * FROM order_items WHERE order_id = $1`
	err = s.db.Select(&items, queryItems, order.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting order items: %w", err)
	}

	cartView := &core.CartView{
		Order: *order,
		Items: []core.CartItemView{},
	}

	if len(items) == 0 {
		return cartView, nil
	}

	// 3. Get product details for each item
	// This is not the most efficient way (N+1 problem), but it's simple for now.
	// For production, a single JOIN query would be better.
	var totalAmount int64
	for _, item := range items {
		var product core.Product
		queryProduct := `SELECT * FROM products WHERE id = $1`
		err := s.db.Get(&product, queryProduct, item.ProductID)
		if err != nil {
			// Handle case where product might be deleted but still in cart
			// For now, we'll just skip it
			continue
		}

		// Recalculate price at this stage
		item.PriceAtPurchase = product.Price
		totalAmount += product.Price * int64(item.Quantity)

		cartView.Items = append(cartView.Items, core.CartItemView{
			OrderItem: item,
			Product:   product,
		})
	}

	// 4. Update the total amount of the order
	if order.TotalAmount != totalAmount {
		order.TotalAmount = totalAmount
		queryUpdateTotal := `UPDATE orders SET total_amount = $1 WHERE id = $2`
		_, err := s.db.Exec(queryUpdateTotal, totalAmount, order.ID)
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
	query := `SELECT * FROM order_items WHERE id = $1`
	err := s.db.Get(&item, query, itemID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
