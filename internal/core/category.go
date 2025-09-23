package core

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category represents a product category
type Category struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name     string    `gorm:"not null" json:"name"`
	Products []Product `json:"products,omitempty"` // one-to-many relationship
}

// TableName sets the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// BeforeCreate will set a UUID rather than relying on default value generation.
func (c *Category) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
