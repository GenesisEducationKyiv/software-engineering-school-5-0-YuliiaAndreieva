package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"

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

	req, err := validator.ValidateAndBind[validator.SubscriptionRequest](h.validator, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("Validated subscription request: email=%s, city=%s, frequency=%s", req.Email, req.City, req.Frequency)

	subscriptionURL, err := url.Parse(h.subscriptionServiceURL)
	if err != nil {
		h.logger.Errorf("Failed to parse subscription service URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(subscriptionURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *SubscriptionHandler) Confirm(c *gin.Context) {
	h.logger.Infof("Proxying subscription confirmation request to %s", h.subscriptionServiceURL)

	req, err := validator.ValidateAndBindURI[validator.TokenRequest](h.validator, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("Validated confirmation token: %s", req.Token)

	subscriptionURL, err := url.Parse(h.subscriptionServiceURL)
	if err != nil {
		h.logger.Errorf("Failed to parse subscription service URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(subscriptionURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	h.logger.Infof("Proxying subscription unsubscribe request to %s", h.subscriptionServiceURL)

	req, err := validator.ValidateAndBindURI[validator.TokenRequest](h.validator, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("Validated unsubscribe token: %s", req.Token)

	subscriptionURL, err := url.Parse(h.subscriptionServiceURL)
	if err != nil {
		h.logger.Errorf("Failed to parse subscription service URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(subscriptionURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}
