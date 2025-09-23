package core

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

type Media struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid" json:"product_id"`
	URL       string    `json:"url"`
	Type      MediaType `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}
