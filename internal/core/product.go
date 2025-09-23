package core

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Product represents the main product entity
type Product struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string           `gorm:"not null" json:"name"`
	Description string           `json:"description"`
	CategoryID  uuid.UUID        `gorm:"type:uuid;column:category_id;not null" json:"category_id"`
	Category    Category         `json:"category"`
	Variants    []ProductVariant `json:"variants,omitempty"`
	Images      []ProductImage   `json:"images,omitempty"`
}

// TableName sets the table name for the Product model
func (Product) TableName() string {
	return "products"
}

// ProductVariant represents a specific variant of a product
type ProductVariant struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string          `gorm:"not null" json:"name"`
	Sku       string          `gorm:"unique;not null" json:"sku"`
	Price     decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"price"`
	Stock     int             `gorm:"not null" json:"stock"`
	ProductID uuid.UUID       `gorm:"type:uuid;column:product_id;not null" json:"product_id"`
	Product   Product         `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;" json:"-"`
}

// TableName sets the table name for the ProductVariant model
func (ProductVariant) TableName() string {
	return "product_variants"
}

// ProductImage represents an image associated with a product
type ProductImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ImageURL  string    `gorm:"column:image_url;not null" json:"image_url"`
	IsMain    bool      `gorm:"column:is_main;default:false" json:"is_main"`
	ProductID uuid.UUID `gorm:"type:uuid;column:product_id;not null" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;" json:"-"`
}

// TableName sets the table name for the ProductImage model
func (ProductImage) TableName() string {
	return "product_images"
}

// BeforeCreate hooks for product models
func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

func (pv *ProductVariant) BeforeCreate(tx *gorm.DB) (err error) {
	if pv.ID == uuid.Nil {
		pv.ID = uuid.New()
	}
	return
}

func (pi *ProductImage) BeforeCreate(tx *gorm.DB) (err error) {
	if pi.ID == uuid.Nil {
		pi.ID = uuid.New()
	}
	return
}
