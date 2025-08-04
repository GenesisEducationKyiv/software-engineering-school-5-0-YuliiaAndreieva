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

	reqBody := c.createTokenGenerationRequest(email, expiresIn)
	jsonBody, err := c.marshalRequestBody(reqBody, "token generation")
	if err != nil {
		return "", err
	}

	req, err := c.createHTTPRequest(ctx, "POST", url, jsonBody)
	if err != nil {
		return "", err
	}

	resp, err := c.sendRequest(req, "token generation")
	if err != nil {
		return "", err
	}

	token, err := c.decodeTokenResponse(resp)
	if err != nil {
		return "", err
	}

	c.logger.Infof("Successfully generated token for email: %s", email)
	return token, nil
}

func (c *TokenClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	c.logger.Debugf("Validating token: %s", token)

	url := fmt.Sprintf("%s/validate", c.baseURL)

	reqBody := c.createTokenValidationRequest(token)
	jsonBody, err := c.marshalRequestBody(reqBody, "token validation")
	if err != nil {
		return false, err
	}

	req, err := c.createHTTPRequest(ctx, "POST", url, jsonBody)
	if err != nil {
		return false, err
	}

	resp, err := c.sendRequest(req, "token validation")
	if err != nil {
		return false, err
	}

	valid, err := c.decodeValidationResponse(resp)
	if err != nil {
		return false, err
	}

	c.logger.Debugf("Token validation result: %t", valid)
	return valid, nil
}

func (c *TokenClient) createTokenGenerationRequest(email, expiresIn string) map[string]interface{} {
	return map[string]interface{}{
		"email":      email,
		"expires_in": expiresIn,
	}
}

func (c *TokenClient) createTokenValidationRequest(token string) map[string]interface{} {
	return map[string]interface{}{
		"token": token,
	}
}

func (c *TokenClient) marshalRequestBody(reqBody map[string]interface{}, operation string) ([]byte, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body for %s: %v", operation, err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return jsonBody, nil
}

func (c *TokenClient) createHTTPRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Errorf("Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *TokenClient) sendRequest(req *http.Request, operation string) (*http.Response, error) {
	c.logger.Debugf("Sending %s request to: %s", operation, req.URL.String())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to send %s request: %v", operation, err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer c.closeResponseBody(resp)

	if err := c.validateResponse(resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *TokenClient) closeResponseBody(resp *http.Response) {
	if closeErr := resp.Body.Close(); closeErr != nil {
		c.logger.Warnf("Failed to close response body: %v", closeErr)
	}
}

func (c *TokenClient) validateResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Token service returned status: %d", resp.StatusCode)
		return fmt.Errorf("token service returned status: %d", resp.StatusCode)
	}
	return nil
}

func (c *TokenClient) decodeTokenResponse(resp *http.Response) (string, error) {
	var response struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Errorf("Failed to decode token generation response: %v", err)
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Token, nil
}

func (c *TokenClient) decodeValidationResponse(resp *http.Response) (bool, error) {
	var response struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Errorf("Failed to decode token validation response: %v", err)
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Valid, nil
}
