package http

import (
	"net/http"
	"strings"

	"subscription-service/internal/core/domain"
	"subscription-service/internal/core/ports/in"
	"subscription-service/internal/core/ports/out"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SubscriptionHandler struct {
	subscribeUseCase       in.SubscribeUseCase
	confirmUseCase         in.ConfirmSubscriptionUseCase
	unsubscribeUseCase     in.UnsubscribeUseCase
	listByFrequencyUseCase in.ListByFrequencyUseCase
	logger                 out.Logger
	validate               *validator.Validate
}

func NewSubscriptionHandler(
	subscribeUseCase in.SubscribeUseCase,
	confirmUseCase in.ConfirmSubscriptionUseCase,
	unsubscribeUseCase in.UnsubscribeUseCase,
	listByFrequencyUseCase in.ListByFrequencyUseCase,
	logger out.Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscribeUseCase:       subscribeUseCase,
		confirmUseCase:         confirmUseCase,
		unsubscribeUseCase:     unsubscribeUseCase,
		listByFrequencyUseCase: listByFrequencyUseCase,
		logger:                 logger,
		validate:               validator.New(),
	}
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req domain.SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for subscription: %v", err)
		c.JSON(http.StatusBadRequest, domain.SubscriptionResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate the request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Errorf("Validation failed for subscription request: %v", err)
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				errorMessages = append(errorMessages, err.Field()+" is required")
			case "email":
				errorMessages = append(errorMessages, "Invalid email format")
			}
		}
		c.JSON(http.StatusBadRequest, domain.SubscriptionResponse{
			Success: false,
			Message: "Validation failed: " + strings.Join(errorMessages, ", "),
		})
		return
	}

	result, err := h.subscribeUseCase.Subscribe(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to process subscription: %v", err)
		c.JSON(http.StatusInternalServerError, domain.SubscriptionResponse{
			Success: false,
			Message: "Failed to process subscription: " + err.Error(),
		})
		return
	}

	if result.Success {
		h.logger.Infof("Subscription processed successfully for email: %s", req.Email)
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Subscription failed for email: %s - %s", req.Email, result.Message)
		c.JSON(http.StatusConflict, result)
	}
}

func (h *SubscriptionHandler) Confirm(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		h.logger.Errorf("Empty token provided")
		c.JSON(http.StatusBadRequest, domain.ConfirmResponse{
			Success: false,
			Message: "Token is required",
		})
		return
	}

	result, err := h.confirmUseCase.ConfirmSubscription(c.Request.Context(), token)
	if err != nil {
		h.logger.Errorf("Failed to confirm subscription: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ConfirmResponse{
			Success: false,
			Message: "Failed to confirm subscription: " + err.Error(),
		})
		return
	}

	if result.Success {
		h.logger.Infof("Subscription confirmed successfully for token: %s", token)
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Subscription confirmation failed for token: %s - %s", token, result.Message)
		c.JSON(http.StatusBadRequest, result)
	}
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		h.logger.Errorf("Empty token provided")
		c.JSON(http.StatusBadRequest, domain.UnsubscribeResponse{
			Success: false,
			Message: "Token is required",
		})
		return
	}

	result, err := h.unsubscribeUseCase.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		h.logger.Errorf("Failed to unsubscribe: %v", err)
		c.JSON(http.StatusInternalServerError, domain.UnsubscribeResponse{
			Success: false,
			Message: "Failed to unsubscribe: " + err.Error(),
		})
		return
	}

	if result.Success {
		h.logger.Infof("Unsubscribed successfully for token: %s", token)
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Unsubscribe failed for token: %s - %s", token, result.Message)
		c.JSON(http.StatusBadRequest, result)
	}
}

func (h *SubscriptionHandler) ListByFrequency(c *gin.Context) {
	var req domain.ListSubscriptionsQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for list subscriptions: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.listByFrequencyUseCase.ListByFrequency(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to list subscriptions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list subscriptions: " + err.Error(),
		})
		return
	}

	h.logger.Infof("Listed %d subscriptions for frequency: %s", len(result.Subscriptions), req.Frequency)
	c.JSON(http.StatusOK, result)
}
