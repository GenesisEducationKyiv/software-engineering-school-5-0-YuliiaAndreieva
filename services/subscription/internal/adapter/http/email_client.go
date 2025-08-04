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

func (c *EmailClient) SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) error {
	c.logger.Debugf("Sending confirmation email to: %s for city: %s", req.To, req.City)

	url := fmt.Sprintf("%s/send/confirmation", c.baseURL)

	jsonBody, err := c.createEmailRequestBody(req)
	if err != nil {
		return err
	}

	httpReq, err := c.createHTTPRequest(ctx, "POST", url, jsonBody)
	if err != nil {
		return err
	}

	if err := c.sendRequest(httpReq); err != nil {
		return err
	}

	c.logger.Infof("Successfully sent confirmation email to %s", req.To)
	return nil
}

func (c *EmailClient) createEmailRequestBody(req domain.ConfirmationEmailRequest) ([]byte, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body for confirmation email: %v", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return jsonBody, nil
}

func (c *EmailClient) createHTTPRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Errorf("Failed to create request for confirmation email: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, nil
}

func (c *EmailClient) sendRequest(req *http.Request) error {
	c.logger.Debugf("Sending confirmation email request to: %s", req.URL.String())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to send confirmation email request: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer c.closeResponseBody(resp)

	return c.validateResponse(resp)
}

func (c *EmailClient) closeResponseBody(resp *http.Response) {
	if closeErr := resp.Body.Close(); closeErr != nil {
		c.logger.Warnf("Failed to close response body: %v", closeErr)
	}
}

func (c *EmailClient) validateResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Email service returned status: %d", resp.StatusCode)
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}
	return nil
}
