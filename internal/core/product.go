package core

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name" binding:"required"`
	Description string    `db:"description" json:"description"`
	Price       int64     `db:"price" json:"price" binding:"required,gte=0"`
	Stock       int       `db:"stock" json:"stock" binding:"required,gte=0"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`

	Attributes json.RawMessage `db:"attributes" json:"attributes,omitempty"`
}

type ProductDetail struct {
	Product          // "Embedding" struct Product di sini
	Media   []*Media `json:"media"` // Daftar media yang terkait
}
