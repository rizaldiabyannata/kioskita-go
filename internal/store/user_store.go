package store

import (
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// UserStore menangani semua operasi database yang berkaitan dengan pengguna.
type UserStore struct {
	db *gorm.DB
}

// NewUserStore membuat instance baru dari UserStore.
func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

// Create menyisipkan user baru ke database.
func (s *UserStore) Create(user *core.User) error {
	if user.Role == "" {
		user.Role = "customer"
	}
	return s.db.Create(user).Error
}

// GetByEmail mengambil user dari database berdasarkan email.
func (s *UserStore) GetByEmail(email string) (*core.User, error) {
	var user core.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAll mengambil semua user dari database.
func (s *UserStore) FindAll() ([]core.User, error) {
	var users []core.User
	err := s.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindByID mengambil user dari database berdasarkan ID.
func (s *UserStore) FindByID(id string) (*core.User, error) {
	var user core.User
	err := s.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update memperbarui user di database.
func (s *UserStore) Update(user *core.User) error {
	return s.db.Save(user).Error
}

// Delete menghapus user dari database.
func (s *UserStore) Delete(id string) error {
	return s.db.Delete(&core.User{}, "id = ?", id).Error
}

// Count mengembalikan jumlah total pengguna di database.
func (s *UserStore) Count() (int, error) {
	var count int64
	err := s.db.Model(&core.User{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
