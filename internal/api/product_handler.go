package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"gorm.io/gorm"
)

type ProductHandler struct {
	productStore  *store.ProductStore
	categoryStore *store.CategoryStore
}

func NewProductHandler(pStore *store.ProductStore, cStore *store.CategoryStore) *ProductHandler {
	return &ProductHandler{
		productStore:  pStore,
		categoryStore: cStore,
	}
}

func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Category routes
	categoryRoutes := router.Group("/categories")
	categoryRoutes.Use(AuthMiddleware()) // Assuming category management is protected
	{
		categoryRoutes.POST("/", h.CreateCategory)
		categoryRoutes.GET("/", h.ListCategories)
	}

	// Product routes
	productRoutes := router.Group("/products")
	{
		productRoutes.GET("/", h.ListProducts)
		productRoutes.GET("/:id", h.GetProductByID)

		protected := productRoutes.Group("/")
		protected.Use(AuthMiddleware())
		{
			protected.POST("/", h.CreateProduct)
			// PUT and DELETE can be added here later
		}
	}
}

// --- Category Handlers ---

func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var category core.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category.ID = uuid.New()

	if err := h.categoryStore.Create(&category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryStore.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// --- Product Handlers ---

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product core.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Assign UUIDs
	product.ID = uuid.New()
	for i := range product.Variants {
		product.Variants[i].ID = uuid.New()
	}
	for i := range product.Images {
		product.Images[i].ID = uuid.New()
	}

	if err := h.productStore.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.productStore.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.productStore.GetProductByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
		return
	}

	c.JSON(http.StatusOK, product)
}
