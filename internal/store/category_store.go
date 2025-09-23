package store

import (
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// CategoryStore handles database operations for Categories.
type CategoryStore struct {
	db *gorm.DB
}

// NewCategoryStore creates a new CategoryStore.
func NewCategoryStore(db *gorm.DB) *CategoryStore {
	return &CategoryStore{db: db}
}

// Create creates a new category.
func (s *CategoryStore) Create(category *core.Category) error {
	return s.db.Create(category).Error
}

// GetAll retrieves all categories.
func (s *CategoryStore) GetAll() ([]core.Category, error) {
	var categories []core.Category
	err := s.db.Find(&categories).Error
	return categories, err
}

// GetByID retrieves a category by its ID.
func (s *CategoryStore) GetByID(id uuid.UUID) (*core.Category, error) {
	var category core.Category
	if err := s.db.First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
