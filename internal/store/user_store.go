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
	if user.Role == "" {
		user.Role = "customer"
	}
	query := `INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING created_at`
	return s.db.QueryRowx(query, user.ID, user.Email, user.PasswordHash, user.Role).Scan(&user.CreatedAt)
}

// GetByEmail mengambil user dari database berdasarkan email.
func (s *UserStore) GetByEmail(email string) (*core.User, error) {
	var user core.User
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`
	err := s.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAll mengambil semua user dari database.
func (s *UserStore) FindAll() ([]core.User, error) {
	var users []core.User
	query := `SELECT id, email, role, created_at FROM users`
	err := s.db.Select(&users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindByID mengambil user dari database berdasarkan ID.
func (s *UserStore) FindByID(id string) (*core.User, error) {
	var user core.User
	query := `SELECT id, email, role, created_at FROM users WHERE id = $1`
	err := s.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update memperbarui user di database.
func (s *UserStore) Update(user *core.User) error {
	query := `UPDATE users SET email = $1 WHERE id = $2`
	_, err := s.db.Exec(query, user.Email, user.ID)
	return err
}

// Delete menghapus user dari database.
func (s *UserStore) Delete(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
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
