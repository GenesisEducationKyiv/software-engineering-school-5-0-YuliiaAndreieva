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

	city, err := h.validateCityParameter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestBody, err := h.createWeatherRequest(city)
	if err != nil {
		h.logger.Errorf("Failed to create request body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	responseBody, statusCode, err := h.forwardRequestToWeatherService(requestBody)
	if err != nil {
		h.logger.Errorf("Failed to forward request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weather data"})
		return
	}

	c.Data(statusCode, "application/json", responseBody)
}

func (h *WeatherHandler) validateCityParameter(c *gin.Context) (string, error) {
	city := c.Query("city")
	if city == "" {
		return "", fmt.Errorf("City parameter is required")
	}
	return city, nil
}

func (h *WeatherHandler) createWeatherRequest(city string) ([]byte, error) {
	requestBody := map[string]string{
		"city": city,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	return jsonBody, nil
}

func (h *WeatherHandler) forwardRequestToWeatherService(requestBody []byte) ([]byte, int, error) {
	weatherURL := fmt.Sprintf("%s/weather", h.weatherServiceURL)
	req, err := http.NewRequest("POST", weatherURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request to weather service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, resp.StatusCode, nil
}
