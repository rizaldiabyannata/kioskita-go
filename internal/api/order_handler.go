package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/service"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orderService *service.OrderService
	orderStore   *store.OrderStore // Keep for direct lookups
}

func NewOrderHandler(orderService *service.OrderService, orderStore *store.OrderStore) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		orderStore:   orderStore,
	}
}

func (h *OrderHandler) RegisterRoutes(router *gin.RouterGroup) {
	orderRoutes := router.Group("/orders")
	orderRoutes.Use(AuthMiddleware()) // All order operations should be protected
	{
		orderRoutes.POST("/", h.CreateOrder)
		orderRoutes.GET("/:id", h.GetOrderByID)
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real app, you might get customer ID from the JWT token (c.Get("userID"))
	// For now, the service doesn't require it as the new Order model has no UserID.

	order, err := h.orderService.CreateNewOrder(req)
	if err != nil {
		// The service layer should return specific errors, but for now, a general error is fine.
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.orderStore.GetOrderByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
		return
	}

	c.JSON(http.StatusOK, order)
}
