package core

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Order represents a customer's order
type Order struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	OrderNumber   string          `gorm:"column:order_number;unique;not null" json:"order_number"`
	OrderDate     time.Time       `gorm:"column:order_date;default:CURRENT_TIMESTAMP" json:"order_date"`
	Status        string          `gorm:"not null" json:"status"`
	Subtotal      decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	TotalDiscount decimal.Decimal `gorm:"column:total_discount;type:decimal(12,2);not null" json:"total_discount"`
	TotalFinal    decimal.Decimal `gorm:"column:total_final;type:decimal(12,2);not null" json:"total_final"`
	VoucherID     *uuid.UUID      `gorm:"type:uuid;column:voucher_id" json:"voucher_id"` // Pointer for nullable
	Voucher       *Voucher        `json:"voucher,omitempty"`
	OrderDetails  []OrderDetail   `json:"order_details,omitempty"`
	Payments      []Payment       `json:"payments,omitempty"`
	Delivery      *Delivery       `json:"delivery,omitempty"`
}

// TableName sets the table name for the Order model
func (Order) TableName() string {
	return "orders"
}

// OrderDetail represents a single item within an order
type OrderDetail struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Quantity     int             `gorm:"not null" json:"quantity"`
	PriceAtOrder decimal.Decimal `gorm:"column:price_at_order;type:decimal(12,2);not null" json:"price_at_order"`
	OrderID      uuid.UUID       `gorm:"type:uuid;column:order_id;not null" json:"order_id"`
	VariantID    uuid.UUID       `gorm:"type:uuid;column:variant_id;not null" json:"variant_id"`
	Order        Order           `json:"-"`
	Variant      ProductVariant  `json:"variant"`
}

// TableName sets the table name for the OrderDetail model
func (OrderDetail) TableName() string {
	return "order_details"
}

// BeforeCreate hooks for order models
func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}

func (od *OrderDetail) BeforeCreate(tx *gorm.DB) (err error) {
	if od.ID == uuid.Nil {
		od.ID = uuid.New()
	}
	return
}
