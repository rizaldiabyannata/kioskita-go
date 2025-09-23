package store

import (
	"database/sql"
	"encoding/json"

	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

type ProductStore struct {
	db *gorm.DB
}

func NewProductStore(db *gorm.DB) *ProductStore {
	return &ProductStore{db: db}
}

func (s *ProductStore) Create(product *core.Product) error {

	if product.Attributes == nil {
		product.Attributes = json.RawMessage("{}")
	}

	return s.db.Create(product).Error
}

func (s *ProductStore) GetByID(id string) (*core.Product, error) {
	var product core.Product
	err := s.db.Preload("Media").First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *ProductStore) List() ([]core.Product, error) {
	var products []core.Product
	err := s.db.Preload("Media").Order("created_at DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *ProductStore) Update(id string, product *core.Product) error {
	if product.Attributes == nil {
		product.Attributes = json.RawMessage("{}")
	}

	result := s.db.Model(&core.Product{}).Where("id = ?", id).Updates(product)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *ProductStore) Delete(id string) error {
	result := s.db.Delete(&core.Product{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
