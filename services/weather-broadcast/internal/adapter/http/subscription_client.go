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

type SubscriptionClient struct {
	baseURL string
	client  *http.Client
	logger  out.Logger
}

func NewSubscriptionClient(baseURL string, client *http.Client, logger out.Logger) out.SubscriptionClient {
	return &SubscriptionClient{
		baseURL: baseURL,
		client:  client,
		logger:  logger,
	}
}

func (c *SubscriptionClient) ListByFrequency(ctx context.Context, query domain.ListSubscriptionsQuery) (*domain.SubscriptionList, error) {
	url := fmt.Sprintf("%s/subscriptions/list", c.baseURL)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
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

	var result domain.SubscriptionList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
