package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"gateway/internal/adapter/validator"
	"gateway/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionServiceURL string
	httpClient             out.HTTPClient
	logger                 out.Logger
	validator              *validator.RequestValidator
}

func NewSubscriptionHandler(subscriptionServiceURL string, httpClient out.HTTPClient, logger out.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionServiceURL: subscriptionServiceURL,
		httpClient:             httpClient,
		logger:                 logger,
		validator:              validator.NewRequestValidator(),
	}
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	h.logger.Infof("Proxying subscription request to %s", h.subscriptionServiceURL)

	req, err := validator.ValidateAndBindJSON[validator.SubscriptionRequest](h.validator, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("Validated subscription request: email=%s, city=%s, frequency=%s", req.Email, req.City, req.Frequency)

	subscriptionURL := h.subscriptionServiceURL + "/subscribe"
	jsonBody, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("Failed to marshal request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	httpReq, err := http.NewRequest("POST", subscriptionURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		h.logger.Errorf("Failed to send request to subscription service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process subscription"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Errorf("Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}

func (h *SubscriptionHandler) Confirm(c *gin.Context) {
	h.logger.Infof("Proxying subscription confirmation request to %s", h.subscriptionServiceURL)

	req, err := validator.ValidateAndBindURI[validator.TokenRequest](h.validator, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("Validated confirmation token: %s", req.Token)

	subscriptionURL := h.subscriptionServiceURL + "/confirm/" + req.Token

	httpReq, err := http.NewRequest("GET", subscriptionURL, nil)
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		h.logger.Errorf("Failed to send request to subscription service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm subscription"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Errorf("Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	h.logger.Infof("Proxying subscription unsubscribe request to %s", h.subscriptionServiceURL)

	req, err := validator.ValidateAndBindURI[validator.TokenRequest](h.validator, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("Validated unsubscribe token: %s", req.Token)

	subscriptionURL := h.subscriptionServiceURL + "/unsubscribe/" + req.Token

	httpReq, err := http.NewRequest("GET", subscriptionURL, nil)
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		h.logger.Errorf("Failed to send request to subscription service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Errorf("Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}
