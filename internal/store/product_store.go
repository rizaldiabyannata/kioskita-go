package store

import (
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// ProductStore handles database operations for Products, Variants, and Images.
type ProductStore struct {
	db *gorm.DB
}

// NewProductStore creates a new ProductStore.
func NewProductStore(db *gorm.DB) *ProductStore {
	return &ProductStore{db: db}
}

// CreateProduct creates a new product along with its variants and images in a transaction.
func (s *ProductStore) CreateProduct(product *core.Product) error {
	return s.db.Create(product).Error
}

// GetAllProducts retrieves all products with their category, variants, and main image preloaded.
func (s *ProductStore) GetAllProducts() ([]core.Product, error) {
	var products []core.Product
	err := s.db.Preload("Category").Preload("Variants").Preload("Images", "is_main = ?", true).Find(&products).Error
	return products, err
}

// GetProductByID retrieves a single product by its ID with all relations preloaded.
func (s *ProductStore) GetProductByID(id uuid.UUID) (*core.Product, error) {
	var product core.Product
	err := s.db.Preload("Category").Preload("Variants").Preload("Images").First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetVariantByID retrieves a product variant by its ID.
func (s *ProductStore) GetVariantByID(id uuid.UUID) (*core.ProductVariant, error) {
	var variant core.ProductVariant
	if err := s.db.First(&variant, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}
