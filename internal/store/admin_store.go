package store

import (
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// AdminStore handles database operations for Admins.
type AdminStore struct {
	db *gorm.DB
}

// NewAdminStore creates a new AdminStore.
func NewAdminStore(db *gorm.DB) *AdminStore {
	return &AdminStore{db: db}
}

// CreateAdmin creates a new admin user.
func (s *AdminStore) CreateAdmin(admin *core.Admin) error {
	return s.db.Create(admin).Error
}

// FindAdminByEmail finds an admin by their email address.
func (s *AdminStore) FindAdminByEmail(email string) (*core.Admin, error) {
	var admin core.Admin
	if err := s.db.Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// FindAdminByID finds an admin by their ID.
func (s *AdminStore) FindAdminByID(id uuid.UUID) (*core.Admin, error) {
	var admin core.Admin
	if err := s.db.Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// Count returns the total number of admins.
func (s *AdminStore) Count() (int64, error) {
	var count int64
	err := s.db.Model(&core.Admin{}).Count(&count).Error
	return count, err
}
