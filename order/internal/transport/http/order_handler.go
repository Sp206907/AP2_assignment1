package http

import (
	"fmt"
	"net/http"
	"order/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/recent", h.GetRecentOrders)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req struct {
		CustomerID string `json:"customer_id" binding:"required"`
		ItemName   string `json:"item_name" binding:"required"`
		Amount     int64  `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("BIND ERROR:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	fmt.Println("KEY:", idempotencyKey)

	order, err := h.uc.CreateOrder(req.CustomerID, req.ItemName, req.Amount, idempotencyKey)
	if err != nil {
		fmt.Println("USECASE ERROR:", err)
		if err.Error() == "payment service unavailable" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	order, err := h.uc.GetOrder(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetRecentOrders(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	orders, err := h.uc.GetRecentOrders(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	err := h.uc.CancelOrder(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
}
