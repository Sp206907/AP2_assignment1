package http

import (
	"net/http"
	"payment/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	uc *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/payments", h.Authorize)
	r.GET("/payments/:order_id", h.GetByOrderID)
}

func (h *PaymentHandler) Authorize(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
		Amount  int64  `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.uc.Authorize(req.OrderID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":         payment.Status,
		"transaction_id": payment.TransactionID,
	})
}

func (h *PaymentHandler) GetByOrderID(c *gin.Context) {
	payment, err := h.uc.GetByOrderID(c.Param("order_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, payment)
}
