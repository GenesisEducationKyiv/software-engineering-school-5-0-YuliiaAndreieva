package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"gateway/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionServiceURL string
	httpClient             out.HTTPClient
	logger                 out.Logger
}

func NewSubscriptionHandler(subscriptionServiceURL string, httpClient out.HTTPClient, logger out.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionServiceURL: subscriptionServiceURL,
		httpClient:             httpClient,
		logger:                 logger,
	}
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	h.logger.Infof("Proxying subscription request to %s", h.subscriptionServiceURL)

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

	subscriptionURL, err := url.Parse(h.subscriptionServiceURL)
	if err != nil {
		h.logger.Errorf("Failed to parse subscription service URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(subscriptionURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}
