package api

import (
	"database/sql"
	"net/http"

	"github.com/rizaldiabyannata/kioskita-go/internal/core"  // Ganti 'kioskita' dengan nama modul Go Anda jika berbeda
	"github.com/rizaldiabyannata/kioskita-go/internal/store" // Ganti 'kioskita' dengan nama modul Go Anda jika berbeda

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // <-- Tambahkan import ini
)

// ProductHandler menangani permintaan HTTP untuk produk.
type ProductHandler struct {
	store *store.ProductStore
}

// NewProductHandler membuat instance baru dari ProductHandler.
func NewProductHandler(store *store.ProductStore) *ProductHandler {
	return &ProductHandler{store: store}
}

// RegisterRoutes mendaftarkan semua rute produk ke router Gin.
func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	productRoutes := router.Group("/products")
	{
		productRoutes.POST("/", h.CreateProduct)
		productRoutes.GET("/", h.ListProducts)
		productRoutes.GET("/:id", h.GetProductByID)
		productRoutes.PUT("/:id", h.UpdateProduct)
		productRoutes.DELETE("/:id", h.DeleteProduct)
	}
}

// CreateProduct menangani pembuatan produk baru.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var newProduct core.Product

	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// === PERUBAHAN UTAMA DI SINI ===
	// Buat ID unik baru sebelum menyimpan ke database.
	newProduct.ID = uuid.NewString()
	// ===============================

	if err := h.store.Create(&newProduct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan produk"})
		return
	}

	c.JSON(http.StatusCreated, newProduct)
}

// GetProductByID menangani pengambilan satu produk berdasarkan ID.
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	product, err := h.store.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil produk"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// ListProducts menangani pengambilan semua produk.
func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar produk"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// UpdateProduct menangani pembaruan produk.
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var updatedProduct core.Product

	if err := c.ShouldBindJSON(&updatedProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.store.Update(id, &updatedProduct)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan untuk diperbarui"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil diperbarui"})
}

// DeleteProduct menangani penghapusan produk.
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := h.store.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan untuk dihapus"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus produk"})
		return
	}

	c.Status(http.StatusNoContent)
}
