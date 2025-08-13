package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
)

type ProductStore struct {
	db *sqlx.DB
}

func NewProductStore(db *sqlx.DB) *ProductStore {
	return &ProductStore{db: db}
}

func (s *ProductStore) Create(product *core.Product) error {

	if product.Attributes == nil {
		product.Attributes = json.RawMessage("{}")
	}

	query := `INSERT INTO products (id, name, description, price, stock, attributes) 
              VALUES ($1, $2, $3, $4, $5, $6) 
              RETURNING created_at`

	return s.db.QueryRowx(query, product.ID, product.Name, product.Description, product.Price, product.Stock, product.Attributes).Scan(&product.CreatedAt)
}

func (s *ProductStore) GetByID(id string) (*core.Product, error) {
	var product core.Product

	query := `SELECT id, name, description, price, stock, created_at, attributes FROM products WHERE id = $1`

	err := s.db.Get(&product, query, id)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *ProductStore) List() ([]core.Product, error) {
	var products []core.Product

	query := `SELECT id, name, description, price, stock, created_at, attributes FROM products ORDER BY created_at DESC`

	err := s.db.Select(&products, query)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *ProductStore) Update(id string, product *core.Product) error {
	if product.Attributes == nil {
		product.Attributes = json.RawMessage("{}")
	}

	query := `UPDATE products 
              SET name = $1, description = $2, price = $3, stock = $4, attributes = $5
              WHERE id = $6`

	result, err := s.db.Exec(query, product.Name, product.Description, product.Price, product.Stock, product.Attributes, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *ProductStore) Delete(id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateStockTx mengurasi stok produk di dalam sebuah transaksi database
func (s *ProductStore) UpdateStockTx(tx *sqlx.Tx, productID string, quantityToDecrease int) error {
	// Ambil produk dengan "FOR UPDATE" untuk mengunci baris dan mencegah race condition
	var currentStock int
	err := tx.Get(&currentStock, "SELECT stock FROM products WHERE id = $1 FOR UPDATE", productID)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan stok produk: %w", err)
	}

	if currentStock < quantityToDecrease {
		return fmt.Errorf("stok tidak cukup untuk produk ID %s", productID)
	}

	newStock := currentStock - quantityToDecrease
	_, err = tx.Exec("UPDATE products SET stock = $1 WHERE id = $2", newStock, productID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui stok produk: %w", err)
	}

	return nil
}
