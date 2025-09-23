package core

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Delivery represents the delivery details for an order
type Delivery struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Address        string          `gorm:"not null" json:"address"`
	RecipientName  string          `gorm:"column:recipient_name;not null" json:"recipient_name"`
	RecipientPhone string          `gorm:"column:recipient_phone;not null" json:"recipient_phone"`
	DriverName     *string         `gorm:"column:driver_name" json:"driver_name"` // Pointer for nullable
	DeliveryFee    decimal.Decimal `gorm:"column:delivery_fee;type:decimal(12,2);not null" json:"delivery_fee"`
	Status         string          `gorm:"not null" json:"status"`
	OrderID        uuid.UUID       `gorm:"type:uuid;unique;column:order_id;not null" json:"order_id"` // 1-to-1
	Order          Order           `json:"-"`
}

// TableName sets the table name for the Delivery model
func (Delivery) TableName() string {
	return "deliveries"
}

// BeforeCreate will set a UUID rather than relying on default value generation.
func (d *Delivery) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return
}
