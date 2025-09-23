package core

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DiscountType defines the type of discount
type DiscountType string

const (
	PercentageDiscount  DiscountType = "PERCENTAGE"
	FixedAmountDiscount DiscountType = "FIXED_AMOUNT"
)

// Voucher represents a discount voucher
type Voucher struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Code          string          `gorm:"unique;not null" json:"code"`
	DiscountType  DiscountType    `gorm:"column:discount_type;type:varchar(20);not null" json:"discount_type"`
	DiscountValue decimal.Decimal `gorm:"column:discount_value;type:decimal(12,2);not null" json:"discount_value"`
	ValidUntil    time.Time       `gorm:"column:valid_until;not null" json:"valid_until"`
	Stock         int             `gorm:"not null" json:"stock"`
	Orders        []Order         `json:"-"` // A voucher can be used in many orders
}

// TableName sets the table name for the Voucher model
func (Voucher) TableName() string {
	return "vouchers"
}

// BeforeCreate will set a UUID rather than relying on default value generation.
func (v *Voucher) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return
}
