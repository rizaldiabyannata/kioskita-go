package core

import "time"

// User merepresentasikan data pengguna di dalam sistem.
type User struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` // Tanda json:"-" menyembunyikan field ini dari respons JSON
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type LoginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
