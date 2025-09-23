package core

import (
	"time"

	"github.com/google/uuid"
)

// User merepresentasikan data pengguna di dalam sistem.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	Email        string    `gorm:"unique" json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
