package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
)

// ProductHandler sekarang memiliki field untuk store dan config.
type ProductHandler struct {
	store  *store.ProductStore
	config *core.StoreConfig
}

// --- PERBAIKAN UTAMA DI SINI ---
// Constructor (fungsi New...) sekarang menerima dua argumen: store dan config.
// Ini akan menyelesaikan error "too many arguments".
func NewProductHandler(store *store.ProductStore, config *core.StoreConfig) *ProductHandler {
	return &ProductHandler{store: store, config: config}
}

// RegisterRoutes mendaftarkan semua rute produk ke router Gin.
func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	productRoutes := router.Group("/products")
	{
		productRoutes.GET("/", h.ListProducts)
		productRoutes.GET("/:id", h.GetProductByID)

		protected := productRoutes.Group("/")
		protected.Use(AuthMiddleware())
		{
			protected.POST("/", h.CreateProduct)
			protected.PUT("/:id", h.UpdateProduct)
			protected.DELETE("/:id", h.DeleteProduct)
		}
	}
}

// CreateProduct menggunakan config untuk validasi dinamis.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var newProduct core.Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi dinamis berdasarkan skema dari config
	var attributes map[string]interface{}
	if err := json.Unmarshal(newProduct.Attributes, &attributes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON pada atribut tidak valid."})
		return
	}

	for _, schema := range h.config.ProductSchema {
		if schema.Required {
			if _, ok := attributes[schema.Key]; !ok {
				errorMsg := fmt.Sprintf("Atribut '%s' (%s) wajib diisi.", schema.Label, schema.Key)
				c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
				return
			}
		}
	}

	newProduct.ID = uuid.NewString()
	if err := h.store.Create(&newProduct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan produk"})
		return
	}
	c.JSON(http.StatusCreated, newProduct)
}

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

func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar produk"})
		return
	}
	c.JSON(http.StatusOK, products)
}

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
