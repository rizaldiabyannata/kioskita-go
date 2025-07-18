package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
)

type ProductHandler struct {
	store      *store.ProductStore
	mediaStore *store.MediaStore
	config     *core.StoreConfig
}

func NewProductHandler(store *store.ProductStore, mediaStore *store.MediaStore, config *core.StoreConfig) *ProductHandler {
	return &ProductHandler{store: store, mediaStore: mediaStore, config: config}
}

func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	productRoutes := router.Group("/products")
	{
		productRoutes.GET("/", h.ListProducts)
		productRoutes.GET("/:id", h.GetProductByID)

		protected := productRoutes.Group("/")
		protected.Use(AuthMiddleware())
		{
			protected.POST("/", h.CreateProductWithMedia) // Menggunakan fungsi baru
			protected.PUT("/:id", h.UpdateProduct)
			protected.DELETE("/:id", h.DeleteProduct)
			protected.POST("/:id/upload", h.UploadMedia) // Endpoint ini tetap ada untuk menambah media nanti
		}
	}
}

// CreateProductWithMedia handles creating a new product with media in a single request.
func (h *ProductHandler) CreateProductWithMedia(c *gin.Context) {
	// 1. Parsing text data from multipart form
	price, err := strconv.ParseInt(c.PostForm("price"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price"})
		return
	}
	stock, err := strconv.Atoi(c.PostForm("stock"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stock"})
		return
	}
	attributesJSON := c.PostForm("attributes")
	var attributes json.RawMessage
	if attributesJSON != "" {
		attributes = json.RawMessage(attributesJSON)
	} else {
		attributes = json.RawMessage("{}")
	}

	// 2. Create a new product object
	newProduct := core.Product{
		ID:          uuid.New(),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Price:       price,
		Stock:       stock,
		Attributes:  attributes,
	}

	// 3. Dynamic attribute validation based on schema
	var attrMap map[string]interface{}
	if err := json.Unmarshal(newProduct.Attributes, &attrMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format for attributes."})
		return
	}
	for _, schema := range h.config.ProductSchema {
		if schema.Required {
			if _, ok := attrMap[schema.Key]; !ok {
				errorMsg := fmt.Sprintf("Attribute '%s' (%s) is required.", schema.Label, schema.Key)
				c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
				return
			}
		}
	}

	// 4. Save the product to the database
	if err := h.store.Create(&newProduct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product"})
		return
	}

	// 5. Process uploaded media files
	form, err := c.MultipartForm()
	if err != nil && err != http.ErrNotMultipart {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to process upload form"})
		return
	}

	var savedMedia []*core.Media
	if form != nil {
		// Process images
		images := form.File["images"]
		if len(images) > 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A maximum of 5 images can be uploaded."})
			return
		}
		imageMedia, err := h.saveFiles(c, newProduct.ID.String(), images, core.MediaTypeImage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		savedMedia = append(savedMedia, imageMedia...)

		// Process video
		videos := form.File["video"]
		if len(videos) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only one video can be uploaded."})
			return
		}
		if len(videos) > 0 {
			videoMedia, err := h.saveFiles(c, newProduct.ID.String(), videos, core.MediaTypeVideo)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			savedMedia = append(savedMedia, videoMedia...)
		}
	}

	// 6. Send success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Product and media created successfully",
		"product": newProduct,
		"media":   savedMedia,
	})
}

// saveFiles is a helper function to save files and record them in the DB.
func (h *ProductHandler) saveFiles(c *gin.Context, productID string, files []*multipart.FileHeader, mediaType core.MediaType) ([]*core.Media, error) {
	var savedMedia []*core.Media
	uploadDir := fmt.Sprintf("./uploads/products/%s", productID)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory")
	}

	for _, file := range files {
		fileName := filepath.Base(file.Filename)
		dst := filepath.Join(uploadDir, fileName)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			return nil, fmt.Errorf("failed to save file: %v", err)
		}

		newMedia := &core.Media{
			ID:        uuid.New(),
			ProductID: uuid.MustParse(productID),
			URL:       dst,
			Type:      mediaType,
		}

		if err := h.mediaStore.Create(newMedia); err != nil {
			return nil, fmt.Errorf("failed to save media information to the database: %v", err)
		}
		savedMedia = append(savedMedia, newMedia)
	}
	return savedMedia, nil
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	convertToPointerSlice := func(media []core.Media) []*core.Media {
		pointerSlice := make([]*core.Media, len(media))
		for i := range media {
			pointerSlice[i] = &media[i]
		}
		return pointerSlice
	}
	id := c.Param("id")

	// 1. Ambil data produk utama dari product store
	product, err := h.store.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
		return
	}

	// 2. Ambil data media yang terkait dari media store
	media, err := h.mediaStore.GetByProductID(id)
	if err != nil {
		// Kita tidak menganggap ini error fatal, mungkin saja produknya belum punya media.
		// Cukup log errornya dan lanjutkan dengan slice media kosong.
		media = []core.Media{} // Pastikan media tidak nil
	}

	// 3. Gabungkan keduanya ke dalam struct respons baru
	productDetail := &core.ProductDetail{
		Product: *product,                     // Salin semua data dari struct product
		Media:   convertToPointerSlice(media), // Convert slice to pointer slice
	}

	// 4. Kirim struct gabungan sebagai respons
	c.JSON(http.StatusOK, productDetail)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product list"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "Product to be updated not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	err := h.store.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product to be deleted not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadMedia is a separate endpoint to add media to an existing product.
func (h *ProductHandler) UploadMedia(c *gin.Context) {
	productID := c.Param("id")
	if _, err := h.store.GetByID(productID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify product"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to process upload form"})
		return
	}

	var savedMedia []*core.Media

	// Process images
	images := form.File["images"]
	if len(images) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A maximum of 5 images can be uploaded."})
		return
	}
	if len(images) > 0 {
		imageMedia, err := h.saveFiles(c, productID, images, core.MediaTypeImage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		savedMedia = append(savedMedia, imageMedia...)
	}

	// Process video
	videos := form.File["video"]
	if len(videos) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only one video can be uploaded."})
		return
	}
	if len(videos) > 0 {
		videoMedia, err := h.saveFiles(c, productID, videos, core.MediaTypeVideo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		savedMedia = append(savedMedia, videoMedia...)
	}

	if len(savedMedia) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files were uploaded with the key 'images' or 'video'."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Files uploaded and saved successfully",
		"product_id": productID,
		"media":      savedMedia,
	})
}
