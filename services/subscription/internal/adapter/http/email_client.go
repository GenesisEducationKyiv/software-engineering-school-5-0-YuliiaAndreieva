package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/out"
)

type EmailClient struct {
	baseURL    string
	httpClient *http.Client
	logger     out.Logger
}

func NewEmailClient(baseURL string, httpClient *http.Client, logger out.Logger) out.EmailService {
	return &EmailClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (c *EmailClient) SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) (out.EmailDeliveryResult, error) {
	c.logger.Debugf("Sending confirmation email to: %s for city: %s", req.To, req.City)

	url := fmt.Sprintf("%s/send/confirmation", c.baseURL)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body for confirmation email: %v", err)
		return out.EmailDeliveryResult{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Errorf("Failed to create request for confirmation email: %v", err)
		return out.EmailDeliveryResult{}, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Sending confirmation email request to: %s", url)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Failed to send confirmation email request: %v", err)
		return out.EmailDeliveryResult{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Email service returned status: %d", resp.StatusCode)
		return out.EmailDeliveryResult{}, fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

	c.logger.Infof("Successfully sent confirmation email to %s", req.To)
	return out.EmailDeliveryResult{
		To:     req.To,
		Status: "delivered",
	}, nil
}
