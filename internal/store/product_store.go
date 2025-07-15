package store

import (
	"database/sql"

	"github.com/rizaldiabyannata/kioskita-go/internal/core" // Ganti 'kioskita' dengan nama modul Go Anda jika berbeda

	"github.com/jmoiron/sqlx"
)

// ProductStore menangani semua operasi database yang berkaitan dengan produk.
type ProductStore struct {
	db *sqlx.DB
}

// NewProductStore membuat instance baru dari ProductStore.
func NewProductStore(db *sqlx.DB) *ProductStore {
	return &ProductStore{db: db}
}

// Create menyisipkan produk baru ke dalam database.
// Sekarang kita juga menyisipkan ID yang sudah dibuat oleh aplikasi.
func (s *ProductStore) Create(product *core.Product) error {
	query := `INSERT INTO products (id, name, description, price, stock) 
              VALUES ($1, $2, $3, $4, $5) 
              RETURNING created_at`

	// Kita hanya perlu mendapatkan `created_at` yang dibuat oleh database.
	// ID sudah ada di dalam struct `product`.
	return s.db.QueryRowx(query, product.ID, product.Name, product.Description, product.Price, product.Stock).Scan(&product.CreatedAt)
}

// GetByID mengambil satu produk dari database berdasarkan ID-nya.
func (s *ProductStore) GetByID(id string) (*core.Product, error) {
	var product core.Product
	query := `SELECT id, name, description, price, stock, created_at FROM products WHERE id = $1`

	err := s.db.Get(&product, query, id)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// List mengambil semua produk dari database.
func (s *ProductStore) List() ([]core.Product, error) {
	var products []core.Product
	query := `SELECT id, name, description, price, stock, created_at FROM products ORDER BY created_at DESC`

	err := s.db.Select(&products, query)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// Update memperbarui data produk yang ada di database.
func (s *ProductStore) Update(id string, product *core.Product) error {
	query := `UPDATE products 
              SET name = $1, description = $2, price = $3, stock = $4 
              WHERE id = $5`

	result, err := s.db.Exec(query, product.Name, product.Description, product.Price, product.Stock, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Jika tidak ada baris yang terpengaruh, berarti produk tidak ditemukan.
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete menghapus produk dari database berdasarkan ID-nya.
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
