package core

import (
	"encoding/json"
	"time"
)

type Product struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name" binding:"required"`
	Description string    `db:"description" json:"description"`
	Price       int64     `db:"price" json:"price" binding:"required,gte=0"`
	Stock       int       `db:"stock" json:"stock" binding:"required,gte=0"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`

	Attributes json.RawMessage `db:"attributes" json:"attributes,omitempty"`
}
