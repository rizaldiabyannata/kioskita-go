package store

import (
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

type MediaStore struct {
	db *gorm.DB
}

func NewMediaStore(db *gorm.DB) *MediaStore {
	return &MediaStore{db: db}
}

func (s *MediaStore) Create(media *core.Media) error {
	return s.db.Create(media).Error
}

func (s *MediaStore) GetByProductID(productID string) ([]core.Media, error) {
	var media []core.Media
	err := s.db.Where("product_id = ?", productID).Find(&media).Error
	return media, err
}
