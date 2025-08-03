package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gateway/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	weatherServiceURL string
	httpClient        out.HTTPClient
	logger            out.Logger
}

func NewWeatherHandler(weatherServiceURL string, httpClient out.HTTPClient, logger out.Logger) *WeatherHandler {
	return &WeatherHandler{
		weatherServiceURL: weatherServiceURL,
		httpClient:        httpClient,
		logger:            logger,
	}
}

func (h *WeatherHandler) Get(c *gin.Context) {
	h.logger.Infof("Proxying weather request to %s", h.weatherServiceURL)

	city := c.Query("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City parameter is required"})
		return
	}

	requestBody := map[string]string{
		"city": city,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		h.logger.Errorf("Failed to marshal request body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	weatherURL := fmt.Sprintf("%s/weather", h.weatherServiceURL)
	req, err := http.NewRequest("POST", weatherURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Errorf("Failed to send request to weather service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weather data"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Errorf("Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}
