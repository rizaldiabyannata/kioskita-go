package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to in-memory sqlite")

	err = store.AutoMigrateDB(db)
	require.NoError(t, err, "Failed to migrate database")

	return db
}

func TestOrderService_CreateNewOrder_Success(t *testing.T) {
	db := setupTestDB(t)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)
	voucherStore := store.NewVoucherStore(db)
	orderService := NewOrderService(db, orderStore, productStore, voucherStore)

	// --- Setup Test Data ---
	// 1. Create a product with variants
	product := &core.Product{
		ID:         uuid.New(),
		Name:       "Kopi Arabika",
		CategoryID: uuid.New(), // Dummy category ID
		Variants: []core.ProductVariant{
			{ID: uuid.New(), Name: "250g", Sku: "KPA-250", Price: decimal.NewFromInt(50000), Stock: 10},
			{ID: uuid.New(), Name: "500g", Sku: "KPA-500", Price: decimal.NewFromInt(90000), Stock: 5},
		},
	}
	require.NoError(t, productStore.CreateProduct(product))

	// --- Test Execution ---
	// 2. Create an order request
	req := CreateOrderRequest{
		Items: []OrderItemRequest{
			{VariantID: product.Variants[0].ID, Quantity: 2}, // 2 * 50000 = 100000
			{VariantID: product.Variants[1].ID, Quantity: 1}, // 1 * 90000 = 90000
		},
	}

	// 3. Call the service method
	order, err := orderService.CreateNewOrder(req)
	require.NoError(t, err)
	require.NotNil(t, order)

	// --- Assertions ---
	// 4. Verify the order totals
	expectedSubtotal := decimal.NewFromInt(190000)
	assert.True(t, expectedSubtotal.Equal(order.Subtotal), "Subtotal should be 190000")
	assert.True(t, decimal.NewFromInt(0).Equal(order.TotalDiscount), "Discount should be 0")
	assert.True(t, expectedSubtotal.Equal(order.TotalFinal), "Final total should be 190000")
	assert.Equal(t, 2, len(order.OrderDetails), "Should have 2 order detail items")

	// 5. Verify stock was decremented
	variant1, _ := productStore.GetVariantByID(product.Variants[0].ID)
	assert.Equal(t, 8, variant1.Stock, "Stock for variant 1 should be decremented by 2")
	variant2, _ := productStore.GetVariantByID(product.Variants[1].ID)
	assert.Equal(t, 4, variant2.Stock, "Stock for variant 2 should be decremented by 1")
}

func TestOrderService_CreateNewOrder_WithPercentageVoucher(t *testing.T) {
	db := setupTestDB(t)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)
	voucherStore := store.NewVoucherStore(db)
	orderService := NewOrderService(db, orderStore, productStore, voucherStore)

	// --- Setup Test Data ---
	product := &core.Product{
		ID:         uuid.New(),
		Name:       "Kopi Robusta",
		CategoryID: uuid.New(),
		Variants: []core.ProductVariant{
			{ID: uuid.New(), Name: "1kg", Sku: "KPR-1000", Price: decimal.NewFromInt(100000), Stock: 20},
		},
	}
	require.NoError(t, productStore.CreateProduct(product))

	voucher := &core.Voucher{
		ID:            uuid.New(),
		Code:          "DISKON10",
		DiscountType:  core.PercentageDiscount,
		DiscountValue: decimal.NewFromInt(10), // 10%
		ValidUntil:    time.Now().Add(24 * time.Hour),
		Stock:         5,
	}
	require.NoError(t, voucherStore.Create(voucher))

	// --- Test Execution ---
	req := CreateOrderRequest{
		Items:       []OrderItemRequest{{VariantID: product.Variants[0].ID, Quantity: 1}}, // Subtotal = 100000
		VoucherCode: &voucher.Code,
	}
	order, err := orderService.CreateNewOrder(req)
	require.NoError(t, err)

	// --- Assertions ---
	expectedSubtotal := decimal.NewFromInt(100000)
	expectedDiscount := decimal.NewFromInt(10000) // 10% of 100000
	expectedFinal := decimal.NewFromInt(90000)
	assert.True(t, expectedSubtotal.Equal(order.Subtotal))
	assert.True(t, expectedDiscount.Equal(order.TotalDiscount), "Discount should be 10000")
	assert.True(t, expectedFinal.Equal(order.TotalFinal), "Final total should be 90000")

	// Verify voucher stock was decremented
	updatedVoucher, _ := voucherStore.FindByCode(voucher.Code)
	assert.Equal(t, 4, updatedVoucher.Stock, "Voucher stock should be decremented by 1")
}

func TestOrderService_CreateNewOrder_WithFixedAmountVoucher(t *testing.T) {
	db := setupTestDB(t)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)
	voucherStore := store.NewVoucherStore(db)
	orderService := NewOrderService(db, orderStore, productStore, voucherStore)

	// --- Setup Test Data ---
	product := &core.Product{
		ID:         uuid.New(),
		Name:       "Teh Melati",
		CategoryID: uuid.New(),
		Variants: []core.ProductVariant{
			{ID: uuid.New(), Name: "1 box", Sku: "TM-1", Price: decimal.NewFromInt(25000), Stock: 30},
		},
	}
	require.NoError(t, productStore.CreateProduct(product))

	voucher := &core.Voucher{
		ID:            uuid.New(),
		Code:          "POTONG5000",
		DiscountType:  core.FixedAmountDiscount,
		DiscountValue: decimal.NewFromInt(5000), // 5000 fixed
		ValidUntil:    time.Now().Add(24 * time.Hour),
		Stock:         10,
	}
	require.NoError(t, voucherStore.Create(voucher))

	// --- Test Execution ---
	req := CreateOrderRequest{
		Items:       []OrderItemRequest{{VariantID: product.Variants[0].ID, Quantity: 2}}, // Subtotal = 50000
		VoucherCode: &voucher.Code,
	}
	order, err := orderService.CreateNewOrder(req)
	require.NoError(t, err)

	// --- Assertions ---
	expectedSubtotal := decimal.NewFromInt(50000)
	expectedDiscount := decimal.NewFromInt(5000)
	expectedFinal := decimal.NewFromInt(45000)
	assert.True(t, expectedSubtotal.Equal(order.Subtotal))
	assert.True(t, expectedDiscount.Equal(order.TotalDiscount), "Discount should be 5000")
	assert.True(t, expectedFinal.Equal(order.TotalFinal), "Final total should be 45000")

	// Verify voucher stock was decremented
	updatedVoucher, _ := voucherStore.FindByCode(voucher.Code)
	assert.Equal(t, 9, updatedVoucher.Stock, "Voucher stock should be decremented by 1")
}

func TestOrderService_CreateNewOrder_InsufficientStock(t *testing.T) {
	db := setupTestDB(t)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)
	voucherStore := store.NewVoucherStore(db)
	orderService := NewOrderService(db, orderStore, productStore, voucherStore)

	product := &core.Product{
		ID:         uuid.New(),
		Name:       "Limited Edition Shirt",
		CategoryID: uuid.New(),
		Variants: []core.ProductVariant{
			{ID: uuid.New(), Name: "M", Sku: "LES-M", Price: decimal.NewFromInt(250000), Stock: 1},
		},
	}
	require.NoError(t, productStore.CreateProduct(product))

	req := CreateOrderRequest{
		Items: []OrderItemRequest{{VariantID: product.Variants[0].ID, Quantity: 2}}, // Request 2, but only 1 in stock
	}

	_, err := orderService.CreateNewOrder(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not enough stock")
}

func TestOrderService_CreateNewOrder_InvalidVoucher(t *testing.T) {
	db := setupTestDB(t)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)
	voucherStore := store.NewVoucherStore(db)
	orderService := NewOrderService(db, orderStore, productStore, voucherStore)

	product := &core.Product{
		ID:         uuid.New(),
		Name:       "Regular T-Shirt",
		CategoryID: uuid.New(),
		Variants: []core.ProductVariant{
			{ID: uuid.New(), Name: "L", Sku: "RTS-L", Price: decimal.NewFromInt(100000), Stock: 10},
		},
	}
	require.NoError(t, productStore.CreateProduct(product))

	invalidCode := "TIDAKVALID"
	req := CreateOrderRequest{
		Items:       []OrderItemRequest{{VariantID: product.Variants[0].ID, Quantity: 1}},
		VoucherCode: &invalidCode,
	}

	_, err := orderService.CreateNewOrder(req)
	require.Error(t, err)
	assert.EqualError(t, err, "invalid voucher code")
}

func TestOrderService_CreateNewOrder_ExpiredVoucher(t *testing.T) {
	db := setupTestDB(t)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)
	voucherStore := store.NewVoucherStore(db)
	orderService := NewOrderService(db, orderStore, productStore, voucherStore)

	product := &core.Product{
		ID:         uuid.New(),
		Name:       "Expired Voucher Product",
		CategoryID: uuid.New(),
		Variants: []core.ProductVariant{
			{ID: uuid.New(), Name: "S", Sku: "EVP-S", Price: decimal.NewFromInt(100000), Stock: 10},
		},
	}
	require.NoError(t, productStore.CreateProduct(product))

	voucher := &core.Voucher{
		ID:            uuid.New(),
		Code:          "EXPIRED",
		DiscountType:  core.FixedAmountDiscount,
		DiscountValue: decimal.NewFromInt(10000),
		ValidUntil:    time.Now().Add(-24 * time.Hour), // Expired yesterday
		Stock:         10,
	}
	require.NoError(t, voucherStore.Create(voucher))

	req := CreateOrderRequest{
		Items:       []OrderItemRequest{{VariantID: product.Variants[0].ID, Quantity: 1}},
		VoucherCode: &voucher.Code,
	}

	_, err := orderService.CreateNewOrder(req)
	require.Error(t, err)
	assert.EqualError(t, err, "voucher has expired")
}

func TestOrderService_CreateNewOrder_VoucherOutOfStock(t *testing.T) {
	db := setupTestDB(t)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)
	voucherStore := store.NewVoucherStore(db)
	orderService := NewOrderService(db, orderStore, productStore, voucherStore)

	product := &core.Product{
		ID:         uuid.New(),
		Name:       "Out of Stock Voucher Product",
		CategoryID: uuid.New(),
		Variants: []core.ProductVariant{
			{ID: uuid.New(), Name: "XL", Sku: "OOSVP-XL", Price: decimal.NewFromInt(100000), Stock: 10},
		},
	}
	require.NoError(t, productStore.CreateProduct(product))

	voucher := &core.Voucher{
		ID:            uuid.New(),
		Code:          "OUTOFSTOCK",
		DiscountType:  core.FixedAmountDiscount,
		DiscountValue: decimal.NewFromInt(10000),
		ValidUntil:    time.Now().Add(24 * time.Hour),
		Stock:         0, // No stock left
	}
	require.NoError(t, voucherStore.Create(voucher))

	req := CreateOrderRequest{
		Items:       []OrderItemRequest{{VariantID: product.Variants[0].ID, Quantity: 1}},
		VoucherCode: &voucher.Code,
	}

	_, err := orderService.CreateNewOrder(req)
	require.Error(t, err)
	assert.EqualError(t, err, "voucher is out of stock")
}
