package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
)

// UserStore menangani semua operasi database yang berkaitan dengan pengguna.
type UserStore struct {
	db *sqlx.DB
}

// NewUserStore membuat instance baru dari UserStore.
func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{db: db}
}

// Create menyisipkan user baru ke database.
func (s *UserStore) Create(user *core.User) error {
	query := `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3) RETURNING created_at`
	return s.db.QueryRowx(query, user.ID, user.Email, user.PasswordHash).Scan(&user.CreatedAt)
}

// GetByEmail mengambil user dari database berdasarkan email.
func (s *UserStore) GetByEmail(email string) (*core.User, error) {
	var user core.User
	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
	err := s.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Count mengembalikan jumlah total pengguna di database.
func (s *UserStore) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	err := s.db.Get(&count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}
