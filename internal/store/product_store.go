package store

import (
	"database/sql"

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
	result := s.db.Model(&core.Product{}).Where("id = ?", id).Updates(product)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateFields updates only the specified fields on Product.
func (s *ProductStore) UpdateFields(id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	result := s.db.Model(&core.Product{}).Where("id = ?", id).Updates(updates)
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

// --- Subtype helpers ---

func (s *ProductStore) CreateClothing(p *core.ClothingProduct) error {
	return s.db.Create(p).Error
}

func (s *ProductStore) GetClothingByProductID(productID string) (*core.ClothingProduct, error) {
	var cp core.ClothingProduct
	if err := s.db.Where("product_id = ?", productID).First(&cp).Error; err != nil {
		return nil, err
	}
	return &cp, nil
}

func (s *ProductStore) UpdateClothingByProductID(productID string, updates map[string]interface{}) error {
	return s.db.Model(&core.ClothingProduct{}).Where("product_id = ?", productID).Updates(updates).Error
}

func (s *ProductStore) DeleteClothingByProductID(productID string) error {
	return s.db.Where("product_id = ?", productID).Delete(&core.ClothingProduct{}).Error
}

func (s *ProductStore) CreateFood(p *core.FoodProduct) error {
	return s.db.Create(p).Error
}

func (s *ProductStore) GetFoodByProductID(productID string) (*core.FoodProduct, error) {
	var fp core.FoodProduct
	if err := s.db.Where("product_id = ?", productID).First(&fp).Error; err != nil {
		return nil, err
	}
	return &fp, nil
}

func (s *ProductStore) UpdateFoodByProductID(productID string, updates map[string]interface{}) error {
	return s.db.Model(&core.FoodProduct{}).Where("product_id = ?", productID).Updates(updates).Error
}

func (s *ProductStore) DeleteFoodByProductID(productID string) error {
	return s.db.Where("product_id = ?", productID).Delete(&core.FoodProduct{}).Error
}
