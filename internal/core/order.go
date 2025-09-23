package core

import (
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

// ShippingAddress menyimpan alamat pengiriman sebagai kolom SQL normal (bukan JSON/NoSQL)
type ShippingAddress struct {
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	AddressLine1  string `json:"address_line1"`
	AddressLine2  string `json:"address_line2"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

type Order struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;" json:"id"`
	UserID          uuid.UUID       `gorm:"type:uuid" json:"user_id"`
	User            User            `gorm:"foreignKey:UserID"`
	TotalAmount     int64           `json:"total_amount"`
	Status          OrderStatus     `json:"status"`
	ShippingAddress ShippingAddress `gorm:"embedded;embeddedPrefix:ship_" json:"shipping_address"`
	Items           []OrderItem     `gorm:"foreignKey:OrderID"`
	CreatedAt       time.Time       `json:"created_at"`
}

type OrderItem struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	OrderID         uuid.UUID `gorm:"type:uuid" json:"order_id"`
	ProductID       uuid.UUID `gorm:"type:uuid" json:"product_id"`
	Product         Product   `gorm:"foreignKey:ProductID"`
	Quantity        int       `json:"quantity"`
	PriceAtPurchase int64     `json:"price_at_purchase"`
	CreatedAt       time.Time `json:"created_at"`
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
