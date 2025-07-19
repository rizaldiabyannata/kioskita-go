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
	orderStore   *store.OrderStore
	productStore *store.ProductStore
}

func NewOrderHandler(orderStore *store.OrderStore, productStore *store.ProductStore) *OrderHandler {
	return &OrderHandler{
		orderStore:   orderStore,
		productStore: productStore,
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

	// Check if product exists and has enough stock
	product, err := h.productStore.GetByID(req.ProductID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	if product.Stock < req.Quantity {
		c.JSON(http.StatusConflict, gin.H{"error": "Not enough stock"})
		return
	}

	// Get or create a cart
	order, err := h.orderStore.GetCartByUserID(userIDUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create a new cart
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

	// Check if item already exists in cart
	existingItem, err := h.orderStore.GetOrderItem(order.ID, req.ProductID)
	if err == nil {
		// Item exists, update quantity
		existingItem.Quantity += req.Quantity
		if product.Stock < existingItem.Quantity {
			c.JSON(http.StatusConflict, gin.H{"error": "Not enough stock for updated quantity"})
			return
		}
		if err := h.orderStore.UpdateOrderItem(existingItem); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item in cart"})
			return
		}
	} else if err == sql.ErrNoRows {
		// Item does not exist, add new item
		newItem := &core.OrderItem{
			ID:              uuid.New(),
			OrderID:         order.ID,
			ProductID:       req.ProductID,
			Quantity:        req.Quantity,
			PriceAtPurchase: product.Price, // Store price at the moment of adding
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

	// Recalculate and respond with the updated cart view
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

	// Verify the item belongs to the user's cart
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

	// Check stock
	product, err := h.productStore.GetByID(item.ProductID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product associated with item not found"})
		return
	}
	if product.Stock < req.Quantity {
		c.JSON(http.StatusConflict, gin.H{"error": "Not enough stock"})
		return
	}

	// Update quantity
	item.Quantity = req.Quantity
	if err := h.orderStore.UpdateOrderItem(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	// Recalculate and respond with the updated cart view
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

	// Verify the item belongs to the user's cart
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

	// Delete the item
	if err := h.orderStore.DeleteOrderItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item from cart"})
		return
	}

	// Recalculate and respond with the updated cart view
	cartView, err := h.orderStore.GetCartViewByUserID(userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated cart"})
		return
	}

	c.JSON(http.StatusOK, cartView)
}
