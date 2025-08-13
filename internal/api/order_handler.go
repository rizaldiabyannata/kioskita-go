package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
)

type OrderHandler struct {
	orderStore *store.OrderStore
	// Hapus productStore dari sini karena sudah di dalam orderStore
}

// Ubah NewOrderHandler, tidak perlu productStore lagi
func NewOrderHandler(orderStore *store.OrderStore) *OrderHandler {
	return &OrderHandler{
		orderStore: orderStore,
	}
}

func (h *OrderHandler) RegisterRoutes(router *gin.RouterGroup) {
	cartRoutes := router.Group("/cart")
	cartRoutes.Use(AuthMiddleware())
	{
		cartRoutes.POST("/items", h.AddItemToCart)
		cartRoutes.GET("/", h.ViewCart)
		cartRoutes.PUT("/items/:item_id", h.UpdateCartItem)
		cartRoutes.DELETE("/items/:item_id", h.DeleteCartItem)
	}

	// Daftarkan grup rute baru untuk checkout
	checkoutRoutes := router.Group("/checkout")
	checkoutRoutes.Use(AuthMiddleware())
	{
		checkoutRoutes.POST("/", h.Checkout)
	}
}

// Handler baru untuk proses checkout
func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID := uuid.MustParse(userID.(string))

	var req core.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Panggil metode checkout dari store
	order, err := h.orderStore.ProcessCheckout(userIDUUID, req.ShippingAddress)
	if err != nil {
		// Cek jenis error untuk memberikan respons yang lebih spesifik
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process checkout: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Checkout successful, order is now pending",
		"order":   order,
	})
}

func (h *OrderHandler) AddItemToCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID := uuid.MustParse(userID.(string))

	var req core.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Dapatkan keranjang atau buat yang baru
	order, err := h.orderStore.GetCartByUserID(userIDUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Buat keranjang baru
			order = &core.Order{
				ID:          uuid.New(),
				UserID:      userIDUUID,
				Status:      core.StatusCart,
				TotalAmount: 0,
				CreatedAt:   time.Now(),
			}
			if err := h.orderStore.CreateOrder(order); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
			return
		}
	}

	// Cek apakah item sudah ada di keranjang
	existingItem, err := h.orderStore.GetOrderItem(order.ID, req.ProductID)
	if err == nil {
		// Item sudah ada, update kuantitas
		existingItem.Quantity += req.Quantity
		if err := h.orderStore.UpdateOrderItem(existingItem); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item in cart"})
			return
		}
	} else if err == sql.ErrNoRows {
		// Item belum ada, tambahkan item baru
		newItem := &core.OrderItem{
			ID:              uuid.New(),
			OrderID:         order.ID,
			ProductID:       req.ProductID,
			Quantity:        req.Quantity,
			PriceAtPurchase: 0, // Harga akan dihitung ulang di GetCartView
			CreatedAt:       time.Now(),
		}
		if err := h.orderStore.AddOrderItem(newItem); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check item in cart"})
		return
	}

	// Tampilkan kembali keranjang yang sudah diperbarui
	cartView, err := h.orderStore.GetCartViewByUserID(userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated cart"})
		return
	}

	c.JSON(http.StatusOK, cartView)
}

func (h *OrderHandler) ViewCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID := uuid.MustParse(userID.(string))

	cartView, err := h.orderStore.GetCartViewByUserID(userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart"})
		return
	}

	if cartView == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Cart is empty"})
		return
	}

	c.JSON(http.StatusOK, cartView)
}

func (h *OrderHandler) UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID := uuid.MustParse(userID.(string))

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req core.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verifikasi item
	item, err := h.orderStore.GetOrderItemByID(itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in cart"})
		return
	}
	cart, err := h.orderStore.GetCartByUserID(userIDUUID)
	if err != nil || item.OrderID != cart.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Item does not belong to user's cart"})
		return
	}

	// Update kuantitas
	item.Quantity = req.Quantity
	if err := h.orderStore.UpdateOrderItem(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	cartView, err := h.orderStore.GetCartViewByUserID(userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated cart"})
		return
	}

	c.JSON(http.StatusOK, cartView)
}

func (h *OrderHandler) DeleteCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID := uuid.MustParse(userID.(string))

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Verifikasi item
	item, err := h.orderStore.GetOrderItemByID(itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in cart"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify item"})
		return
	}
	cart, err := h.orderStore.GetCartByUserID(userIDUUID)
	if err != nil || item.OrderID != cart.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Item does not belong to user's cart"})
		return
	}

	// Hapus item
	if err := h.orderStore.DeleteOrderItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item from cart"})
		return
	}

	cartView, err := h.orderStore.GetCartViewByUserID(userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated cart"})
		return
	}

	c.JSON(http.StatusOK, cartView)
}
