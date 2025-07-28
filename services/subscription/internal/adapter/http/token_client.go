package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"subscription-service/internal/core/ports/out"
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
	url := fmt.Sprintf("%s/generate", c.baseURL)

	reqBody := map[string]interface{}{
		"email":      email,
		"expires_in": expiresIn,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token service returned status: %d", resp.StatusCode)
	}

	var response struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Token, nil
}

func (c *TokenClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	url := fmt.Sprintf("%s/validate", c.baseURL)

	reqBody := map[string]interface{}{
		"token": token,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var response struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Valid, nil
}
