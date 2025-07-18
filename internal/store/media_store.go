package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
)

type MediaStore struct {
	db *sqlx.DB
}

func NewMediaStore(db *sqlx.DB) *MediaStore {
	return &MediaStore{db: db}
}

func (s *MediaStore) Create(media *core.Media) error {
	query := `INSERT INTO media (id, product_id, url, type) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, media.ID, media.ProductID, media.URL, media.Type)
	return err
}

func (s *MediaStore) GetByProductID(productID string) ([]core.Media, error) {
	var media []core.Media
	query := `SELECT id, product_id, url, type, created_at FROM media WHERE product_id = $1`
	err := s.db.Select(&media, query, productID)
	return media, err
}
