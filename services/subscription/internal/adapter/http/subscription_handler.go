package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"

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
	req, err := h.bindAndValidateSubscriptionRequest(c)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.subscribeUseCase.Subscribe(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to process subscription: %v", err)
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to process subscription: "+err.Error())
		return
	}

	h.handleSubscriptionResult(c, *result, req.Email)
}

func (h *SubscriptionHandler) Confirm(c *gin.Context) {
	token, err := h.validateTokenParameter(c)
	if err != nil {
		h.sendConfirmErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.confirmUseCase.ConfirmSubscription(c.Request.Context(), token)
	if err != nil {
		h.logger.Errorf("Failed to confirm subscription: %v", err)
		h.sendConfirmErrorResponse(c, http.StatusInternalServerError, "Failed to confirm subscription: "+err.Error())
		return
	}

	if result.Success {
		h.handleConfirmResult(c, *result, token)
	} else {
		h.sendConfirmErrorResponse(c, http.StatusBadRequest, result.Message)
	}
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token, err := h.validateTokenParameter(c)
	if err != nil {
		h.sendUnsubscribeErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	req := domain.UnsubscribeRequest{Token: token}
	result, err := h.unsubscribeUseCase.Unsubscribe(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to unsubscribe: %v", err)
		h.sendUnsubscribeErrorResponse(c, http.StatusInternalServerError, "Failed to unsubscribe: "+err.Error())
		return
	}

	if result.Success {
		h.handleUnsubscribeResult(c, *result, token)
	} else {
		h.sendUnsubscribeErrorResponse(c, http.StatusBadRequest, result.Message)
	}
}

func (h *SubscriptionHandler) ListByFrequency(c *gin.Context) {
	req, err := h.bindListSubscriptionsRequest(c)
	if err != nil {
		h.sendListErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	result, err := h.listByFrequencyUseCase.ListByFrequency(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to list subscriptions: %v", err)
		h.sendListErrorResponse(c, http.StatusInternalServerError, "Failed to list subscriptions: "+err.Error())
		return
	}

	h.logger.Infof("Listed %d subscriptions for frequency: %s", len(result.Subscriptions), req.Frequency)
	c.JSON(http.StatusOK, result)
}

func (h *SubscriptionHandler) bindAndValidateSubscriptionRequest(c *gin.Context) (domain.SubscriptionRequest, error) {
	var req domain.SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for subscription: %v", err)
		return req, err
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Errorf("Validation failed for subscription request: %v", err)
		return req, h.formatValidationErrors(err)
	}

	return req, nil
}

func (h *SubscriptionHandler) formatValidationErrors(err error) error {
	var errorMessages []string
	for _, validationErr := range err.(validator.ValidationErrors) {
		switch validationErr.Tag() {
		case "required":
			errorMessages = append(errorMessages, validationErr.Field()+" is required")
		case "email":
			errorMessages = append(errorMessages, "Invalid email format")
		}
	}
	return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
}

func (h *SubscriptionHandler) validateTokenParameter(c *gin.Context) (string, error) {
	token := c.Param("token")
	if token == "" {
		h.logger.Errorf("Empty token provided")
		return "", fmt.Errorf("token is required")
	}
	return token, nil
}

func (h *SubscriptionHandler) bindListSubscriptionsRequest(c *gin.Context) (domain.ListSubscriptionsQuery, error) {
	var req domain.ListSubscriptionsQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind JSON for list subscriptions: %v", err)
		return req, err
	}
	return req, nil
}

func (h *SubscriptionHandler) sendErrorResponse(c *gin.Context, statusCode int, message string) {
	response := domain.SubscriptionResponse{
		Success: false,
		Message: message,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.logger.Errorf("Failed to marshal error response: %v", err)
		c.JSON(statusCode, gin.H{"success": false, "message": message})
		return
	}
	if _, err := c.Data(statusCode, "application/json", responseBytes); err != nil {
		h.logger.Errorf("Failed to send error response: %v", err)
		c.JSON(statusCode, gin.H{"success": false, "message": message})
	}
}

func (h *SubscriptionHandler) sendConfirmErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, domain.ConfirmResponse{
		Success: false,
		Message: message,
	})
}

func (h *SubscriptionHandler) sendUnsubscribeErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, domain.UnsubscribeResponse{
		Success: false,
		Message: message,
	})
}

func (h *SubscriptionHandler) sendListErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"message": message,
	})
}

func (h *SubscriptionHandler) handleSubscriptionResult(c *gin.Context, result domain.SubscriptionResponse, email string) {
	if result.Success {
		h.logger.Infof("Subscription processed successfully for email: %s", email)
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Subscription failed for email: %s - %s", email, result.Message)
		c.JSON(http.StatusConflict, result)
	}
}

func (h *SubscriptionHandler) handleConfirmResult(c *gin.Context, result domain.ConfirmResponse, token string) {
	if result.Success {
		h.logger.Infof("Subscription confirmed successfully for token: %s", token)
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Subscription confirmation failed for token: %s - %s", token, result.Message)
		c.JSON(http.StatusBadRequest, result)
	}
}

func (h *SubscriptionHandler) handleUnsubscribeResult(c *gin.Context, result domain.UnsubscribeResponse, token string) {
	if result.Success {
		h.logger.Infof("Unsubscribed successfully for token: %s", token)
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Unsubscribe failed for token: %s - %s", token, result.Message)
		c.JSON(http.StatusBadRequest, result)
	}
}
