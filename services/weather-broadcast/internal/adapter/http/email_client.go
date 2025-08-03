package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/internal/core/ports/out"
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

	url := fmt.Sprintf("%s/send/weather-update", c.baseURL)

	request := domain.WeatherUpdateEmailRequest{
		To:               info.Email,
		Subject:          "Weather Update",
		Name:             "User",
		City:             info.City,
		Description:      info.Weather.Description,
		Temperature:      int(info.Weather.Temperature),
		Humidity:         info.Weather.Humidity,
		WindSpeed:        int(info.Weather.WindSpeed),
		UnsubscribeToken: info.Token,
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Email service returned status: %d for weather update", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.logger.Infof("Successfully sent weather update email to: %s for city: %s", info.Email, info.City)
	return nil
}
