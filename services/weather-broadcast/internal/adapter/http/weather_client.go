package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	sharedlogger "shared/logger"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/internal/core/ports/out"
)

type WeatherClient struct {
	baseURL string
	client  *http.Client
	logger  sharedlogger.Logger
}

func NewWeatherClient(baseURL string, client *http.Client, logger sharedlogger.Logger) out.WeatherClient {
	return &WeatherClient{
		baseURL: baseURL,
		client:  client,
		logger:  logger,
	}
}

func (c *WeatherClient) GetWeatherByCity(ctx context.Context, city string) (*domain.Weather, error) {
	c.logger.Debugf("Getting weather data for city: %s", city)

	url := fmt.Sprintf("%s/weather", c.baseURL)

	request := domain.WeatherRequest{
		City: city,
	}

	body, err := json.Marshal(request)
	if err != nil {
		c.logger.Errorf("Failed to marshal weather request for city %s: %v", city, err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Errorf("Failed to create weather request for city %s: %v", city, err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Sending weather request to: %s for city: %s", url, city)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to make weather request for city %s: %v", city, err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Weather service returned status: %d for city %s", resp.StatusCode, city)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result domain.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Errorf("Failed to decode weather response for city %s: %v", city, err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Infof("Successfully retrieved weather data for city: %s", city)
	return &result.Weather, nil
}

type WeatherRequest struct {
	City string `json:"city"`
}

type WeatherResponse struct {
	Weather domain.Weather `json:"weather"`
}
