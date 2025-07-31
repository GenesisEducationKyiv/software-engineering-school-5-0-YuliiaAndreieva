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
	c.logger.Debugf("Listing subscriptions by frequency: %s, lastID: %d, pageSize: %d", query.Frequency, query.LastID, query.PageSize)

	url := fmt.Sprintf("%s/subscriptions/list", c.baseURL)

	body, err := json.Marshal(query)
	if err != nil {
		c.logger.Errorf("Failed to marshal subscription query: %v", err)
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Errorf("Failed to create subscription request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Sending subscription list request to: %s", url)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to make subscription request: %v", err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Subscription service returned status: %d", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result domain.SubscriptionList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Errorf("Failed to decode subscription response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Infof("Successfully retrieved %d subscriptions for frequency: %s", len(result.Subscriptions), query.Frequency)
	return &result, nil
}
