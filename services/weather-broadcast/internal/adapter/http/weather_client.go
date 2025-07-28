package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"weather-broadcast-service/internal/core/domain"
	"weather-broadcast-service/internal/core/ports/out"
)

type WeatherClient struct {
	baseURL string
	client  *http.Client
	logger  out.Logger
}

func NewWeatherClient(baseURL string, client *http.Client, logger out.Logger) out.WeatherClient {
	return &WeatherClient{
		baseURL: baseURL,
		client:  client,
		logger:  logger,
	}
}

func (c *WeatherClient) GetWeatherByCity(ctx context.Context, city string) (*domain.Weather, error) {
	url := fmt.Sprintf("%s/weather", c.baseURL)

	request := domain.WeatherRequest{
		City: city,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result domain.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Weather, nil
}

type WeatherRequest struct {
	City string `json:"city"`
}

type WeatherResponse struct {
	Weather domain.Weather `json:"weather"`
}
