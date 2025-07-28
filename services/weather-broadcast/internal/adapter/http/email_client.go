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

type EmailClient struct {
	baseURL string
	client  *http.Client
	logger  out.Logger
}

func NewEmailClient(baseURL string, client *http.Client, logger out.Logger) out.WeatherMailer {
	return &EmailClient{
		baseURL: baseURL,
		client:  client,
		logger:  logger,
	}
}

func (c *EmailClient) SendWeather(ctx context.Context, info *domain.WeatherMailSuccessInfo) error {
	url := fmt.Sprintf("%s/email/weather-update", c.baseURL)

	request := domain.WeatherUpdateEmailRequest{
		To:          info.Email,
		Subject:     "Weather Update",
		Name:        "User",
		Location:    info.City,
		Description: info.Weather.Description,
		Temperature: info.Weather.Temperature,
		Humidity:    info.Weather.Humidity,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *EmailClient) SendError(ctx context.Context, info *domain.WeatherMailErrorInfo) error {
	url := fmt.Sprintf("%s/email/weather-error", c.baseURL)

	request := domain.WeatherErrorEmailRequest{
		To:       info.Email,
		Subject:  "Weather Service Error",
		Name:     "User",
		Location: info.City,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type WeatherUpdateEmailRequest struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Temperature int    `json:"temperature"`
	Humidity    int    `json:"humidity"`
}

type WeatherErrorEmailRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Name     string `json:"name"`
	Location string `json:"location"`
}
