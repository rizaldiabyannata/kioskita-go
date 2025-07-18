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
	ID        uuid.UUID `db:"id" json:"id"`
	ProductID uuid.UUID `db:"product_id" json:"product_id"`
	URL       string    `db:"url" json:"url"`
	Type      MediaType `db:"type" json:"type"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
