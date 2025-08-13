package core

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusCart     OrderStatus = "cart"
	StatusPending  OrderStatus = "pending"
	StatusPaid     OrderStatus = "paid"
	StatusShipped  OrderStatus = "shipped"
	StatusComplete OrderStatus = "complete"
	StatusCanceled OrderStatus = "canceled"
)

type Order struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	UserID          uuid.UUID       `db:"user_id" json:"user_id"`
	TotalAmount     int64           `db:"total_amount" json:"total_amount"`
	Status          OrderStatus     `db:"status" json:"status"`
	ShippingAddress json.RawMessage `db:"shipping_address" json:"shipping_address"` // Diubah ke json.RawMessage
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
}

type OrderItem struct {
	ID              uuid.UUID `db:"id" json:"id"`
	OrderID         uuid.UUID `db:"order_id" json:"order_id"`
	ProductID       uuid.UUID `db:"product_id" json:"product_id"`
	Quantity        int       `db:"quantity" json:"quantity"`
	PriceAtPurchase int64     `db:"price_at_purchase" json:"price_at_purchase"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type CartView struct {
	Order
	Items []CartItemView `json:"items"`
}

type CartItemView struct {
	OrderItem
	Product Product `json:"product"`
}

type AddToCartRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

// CheckoutRequest adalah payload yang dikirim oleh user saat checkout
type CheckoutRequest struct {
	ShippingAddress json.RawMessage `json:"shipping_address" binding:"required"`
}
