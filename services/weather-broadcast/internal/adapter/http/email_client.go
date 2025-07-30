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
	c.logger.Debugf("Sending weather update email to: %s for city: %s", info.Email, info.City)

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
		c.logger.Errorf("Failed to marshal weather update email request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Errorf("Failed to create weather update email request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Sending weather update email request to: %s", url)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to make weather update email request: %v", err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Email service returned status: %d for weather update", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.logger.Infof("Successfully sent weather update email to: %s for city: %s", info.Email, info.City)
	return nil
}

func (c *EmailClient) SendError(ctx context.Context, info *domain.WeatherMailErrorInfo) error {
	c.logger.Debugf("Sending error email to: %s for city: %s", info.Email, info.City)

	url := fmt.Sprintf("%s/email/error", c.baseURL)

	request := map[string]interface{}{
		"to":      info.Email,
		"subject": "Weather Update Error",
		"city":    info.City,
		"message": "Unable to retrieve weather data for your city",
	}

	body, err := json.Marshal(request)
	if err != nil {
		c.logger.Errorf("Failed to marshal error email request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Errorf("Failed to create error email request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Sending error email request to: %s", url)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to make error email request: %v", err)
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Email service returned status: %d for error email", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.logger.Infof("Successfully sent error email to: %s for city: %s", info.Email, info.City)
	return nil
}

type WeatherUpdateEmailRequest struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	Name        string `json:"name"`
	City        string `json:"city"`
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
