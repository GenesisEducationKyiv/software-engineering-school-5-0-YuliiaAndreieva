package http

import (
	"errors"
	"log"
	"net/http"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/service"
	"weather-api/internal/handler/http/request"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req request.SubscribeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid subscription request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidInput.Error()})
		return
	}

	log.Printf("Received subscription request for city: %s, frequency: %s", req.City, req.Frequency)

	if req.Frequency != domain.FrequencyDaily && req.Frequency != domain.FrequencyHourly {
		log.Printf("Invalid frequency in subscription request: %s", req.Frequency)
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidInput.Error()})
		return
	}

	_, err := h.subscriptionService.Subscribe(c, req.Email, req.City, req.Frequency)
	if err != nil {
		log.Printf("Failed to process subscription: %v", err)
		switch {
		case errors.Is(err, domain.ErrEmailAlreadySubscribed):
			c.JSON(http.StatusConflict, gin.H{"error": domain.ErrEmailAlreadySubscribed.Error()})
		case errors.Is(err, domain.ErrCityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "City not found"})
		}
		return
	}
	log.Printf("Successfully processed subscription request")
	c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Confirmation email sent."})
}

func (h *SubscriptionHandler) Confirm(c *gin.Context) {
	token := c.Param("token")
	log.Printf("Received confirmation request")

	if err := h.subscriptionService.Confirm(c, token); err != nil {
		log.Printf("Failed to confirm subscription: %v", err)
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidToken.Error()})
		case errors.Is(err, domain.ErrTokenNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrTokenNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	log.Printf("Successfully confirmed subscription")
	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed"})
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	log.Printf("Received unsubscribe request")

	if err := h.subscriptionService.Unsubscribe(c, token); err != nil {
		log.Printf("Failed to unsubscribe: %v", err)
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidToken.Error()})
		case errors.Is(err, domain.ErrTokenNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrTokenNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	log.Printf("Successfully processed unsubscribe request")
	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed"})
}
