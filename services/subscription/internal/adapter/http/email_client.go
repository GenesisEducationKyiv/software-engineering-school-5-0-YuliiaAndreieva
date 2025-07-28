package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"subscription-service/internal/core/ports/out"
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

func (c *EmailClient) SendConfirmationEmail(ctx context.Context, email, city, token string) error {
	url := fmt.Sprintf("%s/send-confirmation", c.baseURL)

	reqBody := map[string]interface{}{
		"email": email,
		"city":  city,
		"token": token,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

	c.logger.Infof("Successfully sent confirmation email to %s", email)
	return nil
}
