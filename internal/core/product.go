package core

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Price       int64     `json:"price" binding:"required,gte=0"`
	Stock       int       `json:"stock" binding:"required,gte=0"`
	CreatedAt   time.Time `json:"created_at"`

	Attributes json.RawMessage `json:"attributes,omitempty"`
	Media      []Media         `gorm:"foreignKey:ProductID" json:"media"` // Daftar media yang terkait
}

type ProductDetail struct {
	Product          // "Embedding" struct Product di sini
	Media   []*Media `json:"media"` // Daftar media yang terkait
}
