package store

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
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

	err = AutoMigrateDB(db)
	require.NoError(t, err, "Failed to migrate database")

	return db
}

func TestAdminStore_CreateAndFind(t *testing.T) {
	db := setupTestDB(t)
	adminStore := NewAdminStore(db)

	newAdmin := &core.Admin{
		Name:  "Test Admin",
		Email: "test@example.com",
	}
	err := newAdmin.HashPassword("password123")
	require.NoError(t, err)

	err = adminStore.CreateAdmin(newAdmin)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, newAdmin.ID)

	// Test FindAdminByEmail
	foundByEmail, err := adminStore.FindAdminByEmail("test@example.com")
	require.NoError(t, err)
	assert.Equal(t, newAdmin.ID, foundByEmail.ID)
	assert.Equal(t, "Test Admin", foundByEmail.Name)

	// Test FindAdminByID
	foundByID, err := adminStore.FindAdminByID(newAdmin.ID)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", foundByID.Email)
}

func TestCategoryStore_CreateAndGetAll(t *testing.T) {
	db := setupTestDB(t)
	categoryStore := NewCategoryStore(db)

	cat1 := &core.Category{Name: "Kopi"}
	cat2 := &core.Category{Name: "Teh"}
	require.NoError(t, categoryStore.Create(cat1))
	require.NoError(t, categoryStore.Create(cat2))

	categories, err := categoryStore.GetAll()
	require.NoError(t, err)
	assert.Len(t, categories, 2)
}

func TestProductStore_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	productStore := NewProductStore(db)
	categoryStore := NewCategoryStore(db)

	// Create a category first
	category := &core.Category{Name: "Pakaian"}
	require.NoError(t, categoryStore.Create(category))

	// Create a complex product
	product := &core.Product{
		Name:        "Kemeja Flanel",
		Description: "Kemeja flanel kotak-kotak",
		CategoryID:  category.ID,
		Variants: []core.ProductVariant{
			{Name: "M", Sku: "FLANEL-M", Price: decimal.NewFromInt(150000), Stock: 10},
			{Name: "L", Sku: "FLANEL-L", Price: decimal.NewFromInt(150000), Stock: 15},
		},
		Images: []core.ProductImage{
			{ImageURL: "/images/flanel1.jpg", IsMain: true},
			{ImageURL: "/images/flanel2.jpg", IsMain: false},
		},
	}
	err := productStore.CreateProduct(product)
	require.NoError(t, err)

	// Retrieve and verify
	foundProduct, err := productStore.GetProductByID(product.ID)
	require.NoError(t, err)
	assert.NotNil(t, foundProduct)
	assert.Equal(t, "Kemeja Flanel", foundProduct.Name)
	assert.Len(t, foundProduct.Variants, 2, "Should have 2 variants")
	assert.Len(t, foundProduct.Images, 2, "Should have 2 images")
	assert.NotNil(t, foundProduct.Category, "Category should be preloaded")
	assert.Equal(t, "Pakaian", foundProduct.Category.Name)
}

func TestVoucherStore_CreateAndFindByCode(t *testing.T) {
	db := setupTestDB(t)
	voucherStore := NewVoucherStore(db)

	voucher := &core.Voucher{
		Code:          "HEMAT20",
		DiscountType:  core.PercentageDiscount,
		DiscountValue: decimal.NewFromInt(20),
		Stock:         100,
	}
	require.NoError(t, voucherStore.Create(voucher))

	foundVoucher, err := voucherStore.FindByCode("HEMAT20")
	require.NoError(t, err)
	assert.Equal(t, voucher.ID, foundVoucher.ID)
	assert.True(t, decimal.NewFromInt(20).Equal(foundVoucher.DiscountValue))
}
