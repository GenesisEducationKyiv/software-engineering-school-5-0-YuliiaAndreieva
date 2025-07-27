package http

import (
	"net/http"

	"subscription-service/internal/core/domain"
	"subscription-service/internal/core/ports/in"
	"subscription-service/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscribeUseCase   in.SubscribeUseCase
	confirmUseCase     in.ConfirmSubscriptionUseCase
	unsubscribeUseCase in.UnsubscribeUseCase
	logger             out.Logger
}

func NewSubscriptionHandler(
	subscribeUseCase in.SubscribeUseCase,
	confirmUseCase in.ConfirmSubscriptionUseCase,
	unsubscribeUseCase in.UnsubscribeUseCase,
	logger out.Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscribeUseCase:   subscribeUseCase,
		confirmUseCase:     confirmUseCase,
		unsubscribeUseCase: unsubscribeUseCase,
		logger:             logger,
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