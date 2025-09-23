package core

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Price       int64     `json:"price" binding:"required,gte=0"`
	Stock       int       `json:"stock" binding:"required,gte=0"`
	CreatedAt   time.Time `json:"created_at"`
	Media       []Media   `gorm:"foreignKey:ProductID" json:"media"` // Daftar media yang terkait
}

type ProductDetail struct {
	Product                   // "Embedding" struct Product di sini
	Media    []*Media         `json:"media"` // Daftar media yang terkait
	Clothing *ClothingProduct `json:"clothing,omitempty"`
	Food     *FoodProduct     `json:"food,omitempty"`
}

// Clothing specific attributes (normalized, no JSON/BSON)
type ClothingProduct struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"product_id"`
	Warna     string    `json:"warna"`
	Ukuran    string    `json:"ukuran"` // e.g., S, M, L, XL, XXL
	Bahan     string    `json:"bahan"`
	CreatedAt time.Time `json:"created_at"`
}

// Food/Drink specific attributes (normalized)
type FoodProduct struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	ProductID   uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"product_id"`
	AsalBiji    string    `json:"asal_biji"`
	LevelGiling string    `json:"level_giling"` // e.g., Biji Utuh, Kasar, Medium, Halus
	CreatedAt   time.Time `json:"created_at"`
}
