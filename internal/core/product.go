package core

import (
	"time"
)

// Product merepresentasikan data sebuah produk di dalam sistem.
// Tanda `db` digunakan oleh sqlx untuk memetakan kolom database ke field struct.
// Tanda `json` digunakan oleh Gin untuk serialisasi/deserialisasi data JSON.
type Product struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name" binding:"required"`
	Description string    `db:"description" json:"description"`
	Price       int64     `db:"price" json:"price" binding:"required,gte=0"`
	Stock       int       `db:"stock" json:"stock" binding:"required,gte=0"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}