package handlers

import (
	"net/http"
	"strconv"

	"chinese-learning-app/internal/services"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

type CreateCheckoutRequest struct {
	Gateway    string `json:"gateway"`
	PlanCode   string `json:"planCode" binding:"required"`
	Currency   string `json:"currency" binding:"required"`
	SuccessURL string `json:"successUrl"`
	CancelURL  string `json:"cancelUrl"`
}

func (h *PaymentHandler) GetCatalog(c *gin.Context) {
	userID, _ := c.Get("userId")
	catalog, err := h.paymentService.GetCatalog(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment catalog"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"catalog": catalog})
}

func (h *PaymentHandler) CreateCheckout(c *gin.Context) {
	userID, _ := c.Get("userId")

	var req CreateCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.paymentService.CreateCheckout(userID.(uint), services.CreateCheckoutRequest{
		Gateway:    req.Gateway,
		PlanCode:   req.PlanCode,
		Currency:   req.Currency,
		SuccessURL: req.SuccessURL,
		CancelURL:  req.CancelURL,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checkout": result})
}

func (h *PaymentHandler) ConfirmCheckout(c *gin.Context) {
	userID, _ := c.Get("userId")
	orderIDValue, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	result, err := h.paymentService.ConfirmCheckout(userID.(uint), uint(orderIDValue))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checkout": result})
}
