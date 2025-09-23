package core

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Payment represents a payment record for an order
type Payment struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	PaymentMethod string          `gorm:"column:payment_method;not null" json:"payment_method"`
	Amount        decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	PaymentDate   time.Time       `gorm:"column:payment_date;not null" json:"payment_date"`
	Status        string          `gorm:"not null" json:"status"`
	OrderID       uuid.UUID       `gorm:"type:uuid;column:order_id;not null" json:"order_id"`
	Order         Order           `json:"-"`
}

// TableName sets the table name for the Payment model
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate will set a UUID rather than relying on default value generation.
func (p *Payment) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
