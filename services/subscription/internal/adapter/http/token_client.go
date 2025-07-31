package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"subscription/internal/core/ports/out"
)

type TokenClient struct {
	baseURL    string
	httpClient *http.Client
	logger     out.Logger
}

func NewTokenClient(baseURL string, httpClient *http.Client, logger out.Logger) out.TokenService {
	return &TokenClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (c *TokenClient) GenerateToken(ctx context.Context, email string, expiresIn string) (string, error) {
	c.logger.Debugf("Generating token for email: %s with expiration: %s", email, expiresIn)

	url := fmt.Sprintf("%s/generate", c.baseURL)

	reqBody := map[string]interface{}{
		"email":      email,
		"expires_in": expiresIn,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body for token generation: %v", err)
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Errorf("Failed to create request for token generation: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Sending token generation request to: %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to send token generation request: %v", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Token service returned status: %d", resp.StatusCode)
		return "", fmt.Errorf("token service returned status: %d", resp.StatusCode)
	}

	var response struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Errorf("Failed to decode token generation response: %v", err)
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Infof("Successfully generated token for email: %s", email)
	return response.Token, nil
}

func (c *TokenClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	c.logger.Debugf("Validating token: %s", token)

	url := fmt.Sprintf("%s/validate", c.baseURL)

	reqBody := map[string]interface{}{
		"token": token,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body for token validation: %v", err)
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Errorf("Failed to create request for token validation: %v", err)
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Sending token validation request to: %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to send token validation request: %v", err)
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Token service returned status: %d", resp.StatusCode)
		return false, fmt.Errorf("token service returned status: %d", resp.StatusCode)
	}

	var response struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Errorf("Failed to decode token validation response: %v", err)
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debugf("Token validation result: %t", response.Valid)
	return response.Valid, nil
}
