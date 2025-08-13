package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
)

type OrderStore struct {
	db           *sqlx.DB
	productStore *ProductStore // Tambahkan productStore
}

// Ubah NewOrderStore untuk menerima ProductStore
func NewOrderStore(db *sqlx.DB, productStore *ProductStore) *OrderStore {
	return &OrderStore{
		db:           db,
		productStore: productStore,
	}
}

// GetCartByUserID finds an active cart for a user.
func (s *OrderStore) GetCartByUserID(userID uuid.UUID) (*core.Order, error) {
	var order core.Order
	query := `SELECT * FROM orders WHERE user_id = $1 AND status = $2`
	err := s.db.Get(&order, query, userID, core.StatusCart)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// CreateOrder creates a new order (cart).
func (s *OrderStore) CreateOrder(order *core.Order) error {
	query := `INSERT INTO orders (id, user_id, total_amount, status, created_at)
              VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(query, order.ID, order.UserID, order.TotalAmount, order.Status, order.CreatedAt)
	return err
}

// AddOrderItem adds a product to an order.
func (s *OrderStore) AddOrderItem(item *core.OrderItem) error {
	query := `INSERT INTO order_items (id, order_id, product_id, quantity, price_at_purchase, created_at)
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.Exec(query, item.ID, item.OrderID, item.ProductID, item.Quantity, item.PriceAtPurchase, item.CreatedAt)
	return err
}

// GetOrderItem finds a specific item in an order.
func (s *OrderStore) GetOrderItem(orderID, productID uuid.UUID) (*core.OrderItem, error) {
	var item core.OrderItem
	query := `SELECT * FROM order_items WHERE order_id = $1 AND product_id = $2`
	err := s.db.Get(&item, query, orderID, productID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateOrderItem updates an item's quantity.
func (s *OrderStore) UpdateOrderItem(item *core.OrderItem) error {
	query := `UPDATE order_items SET quantity = $1 WHERE id = $2`
	_, err := s.db.Exec(query, item.Quantity, item.ID)
	return err
}

// DeleteOrderItem removes an item from an order.
func (s *OrderStore) DeleteOrderItem(itemID uuid.UUID) error {
	query := `DELETE FROM order_items WHERE id = $1`
	res, err := s.db.Exec(query, itemID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetCartViewByUserID retrieves the full cart view with products.
func (s *OrderStore) GetCartViewByUserID(userID uuid.UUID) (*core.CartView, error) {
	// 1. Get the active cart order
	order, err := s.GetCartByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	// 2. Get all items for that order
	var items []core.OrderItem
	queryItems := `SELECT * FROM order_items WHERE order_id = $1`
	err = s.db.Select(&items, queryItems, order.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting order items: %w", err)
	}

	cartView := &core.CartView{
		Order: *order,
		Items: []core.CartItemView{},
	}

	if len(items) == 0 {
		return cartView, nil
	}

	var totalAmount int64
	for _, item := range items {
		var product core.Product
		queryProduct := `SELECT * FROM products WHERE id = $1`
		err := s.db.Get(&product, queryProduct, item.ProductID)
		if err != nil {
			continue
		}
		item.PriceAtPurchase = product.Price
		totalAmount += product.Price * int64(item.Quantity)

		cartView.Items = append(cartView.Items, core.CartItemView{
			OrderItem: item,
			Product:   product,
		})
	}

	if order.TotalAmount != totalAmount {
		order.TotalAmount = totalAmount
		queryUpdateTotal := `UPDATE orders SET total_amount = $1 WHERE id = $2`
		_, err := s.db.Exec(queryUpdateTotal, totalAmount, order.ID)
		if err != nil {
			fmt.Printf("Warning: failed to update total amount for order %s: %v\n", order.ID, err)
		}
		cartView.Order.TotalAmount = totalAmount
	}

	return cartView, nil
}

// GetOrderItemByID retrieves a single order item by its ID.
func (s *OrderStore) GetOrderItemByID(itemID uuid.UUID) (*core.OrderItem, error) {
	var item core.OrderItem
	query := `SELECT * FROM order_items WHERE id = $1`
	err := s.db.Get(&item, query, itemID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ProcessCheckout menangani seluruh logika checkout dalam satu transaksi
func (s *OrderStore) ProcessCheckout(userID uuid.UUID, shippingAddress json.RawMessage) (*core.Order, error) {
	// Memulai transaksi
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	// Pastikan transaksi di-rollback jika ada error
	defer tx.Rollback()

	// 1. Ambil keranjang belanja (cart) yang aktif
	var cart core.Order
	queryCart := `SELECT * FROM orders WHERE user_id = $1 AND status = $2 FOR UPDATE`
	if err := tx.Get(&cart, queryCart, userID, core.StatusCart); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("keranjang belanja tidak ditemukan atau kosong")
		}
		return nil, fmt.Errorf("gagal mendapatkan keranjang belanja: %w", err)
	}

	// 2. Ambil semua item di dalam keranjang
	var items []core.OrderItem
	queryItems := `SELECT * FROM order_items WHERE order_id = $1`
	if err := tx.Select(&items, queryItems, cart.ID); err != nil {
		return nil, fmt.Errorf("gagal mendapatkan item keranjang: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("keranjang belanja kosong, tidak bisa checkout")
	}

	// 3. Validasi stok dan kurangi stok untuk setiap item
	for _, item := range items {
		err := s.productStore.UpdateStockTx(tx, item.ProductID.String(), item.Quantity)
		if err != nil {
			// Jika error (misal, stok tidak cukup), transaksi akan di-rollback
			return nil, fmt.Errorf("gagal memproses item %s: %w", item.ProductID, err)
		}
	}

	// 4. Update status order menjadi 'pending' dan tambahkan alamat pengiriman
	cart.Status = core.StatusPending
	cart.ShippingAddress = shippingAddress
	queryUpdateOrder := `UPDATE orders SET status = $1, shipping_address = $2 WHERE id = $3`
	if _, err := tx.Exec(queryUpdateOrder, cart.Status, cart.ShippingAddress, cart.ID); err != nil {
		return nil, fmt.Errorf("gagal memperbarui status pesanan: %w", err)
	}

	// Jika semua langkah berhasil, commit transaksi
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit transaksi: %w", err)
	}

	return &cart, nil
}
