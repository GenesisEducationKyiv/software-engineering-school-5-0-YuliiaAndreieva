package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	httperrors "weather-api/internal/adapter/handler/http/errors"
	"weather-api/internal/adapter/handler/http/request"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"
)

type SubscriptionHandler struct {
	subscriptionService ports.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService ports.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req request.SubscribeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid subscription request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": httperrors.ErrInvalidInput.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received subscription request for city: %s, frequency: %s", req.City, req.Frequency)

	_, err := h.subscriptionService.Subscribe(c, ports.SubscribeOptions{
		Email:     req.Email,
		City:      req.City,
		Frequency: req.Frequency,
	})
	if err != nil {
		log.Printf("Unable to process subscription: %v", err)
		switch {
		case errors.Is(err, domain.ErrEmailAlreadySubscribed):
			c.JSON(http.StatusConflict, gin.H{"error": httperrors.ErrEmailAlreadySubscribed.Error()})
		case errors.Is(err, domain.ErrCityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": httperrors.ErrCityNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	log.Printf("Successfully processed subscription request")
	c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Confirmation email sent."})
}

func (h *SubscriptionHandler) Confirm(c *gin.Context) {
	token := c.Param("token")
	tokenReq := request.NewTokenRequest(token)

	if err := tokenReq.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received confirmation request")

	if err := h.subscriptionService.Confirm(c, token); err != nil {
		log.Printf("Unable to confirm subscription: %v", err)
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": httperrors.ErrInvalidToken.Error()})
		case errors.Is(err, domain.ErrTokenNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": httperrors.ErrTokenNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	log.Printf("Successfully confirmed subscription")
	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed"})
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	tokenReq := request.NewTokenRequest(token)

	if err := tokenReq.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received unsubscribe request")

	if err := h.subscriptionService.Unsubscribe(c, token); err != nil {
		log.Printf("Unable to unsubscribe: %v", err)
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": httperrors.ErrInvalidToken.Error()})
		case errors.Is(err, domain.ErrTokenNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": httperrors.ErrTokenNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	log.Printf("Successfully processed unsubscribe request")
	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed"})
}
