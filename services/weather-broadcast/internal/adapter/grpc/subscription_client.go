package grpc

import (
	"context"
	"fmt"
	pb "proto/subscription"
	sharedlogger "shared/logger"
	"weather-broadcast/internal/core/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SubscriptionClient struct {
	client pb.SubscriptionServiceClient
	logger sharedlogger.Logger
}

func NewSubscriptionClient(address string, logger sharedlogger.Logger) (*SubscriptionClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to subscription service: %w", err)
	}

	client := pb.NewSubscriptionServiceClient(conn)
	return &SubscriptionClient{
		client: client,
		logger: logger,
	}, nil
}

func (c *SubscriptionClient) ListByFrequency(ctx context.Context, query domain.ListSubscriptionsQuery) (*domain.SubscriptionList, error) {
	c.logger.Debugf("Listing subscriptions for frequency: %s", query.Frequency)

	req := &pb.ListByFrequencyRequest{
		Frequency: string(query.Frequency),
		LastId:    int32(query.LastID),
		PageSize:  int32(query.PageSize),
	}

	resp, err := c.client.ListByFrequency(ctx, req)
	if err != nil {
		c.logger.Errorf("Failed to list subscriptions: %v", err)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	subscriptions := make([]domain.Subscription, len(resp.Subscriptions))
	for i, sub := range resp.Subscriptions {
		subscriptions[i] = domain.Subscription{
			ID:        int(sub.Id),
			Email:     sub.Email,
			City:      sub.City,
			Frequency: domain.Frequency(sub.Frequency),
			Confirmed: sub.Confirmed,
			Token:     sub.Token,
		}
	}

	c.logger.Infof("Successfully retrieved %d subscriptions for frequency: %s", len(subscriptions), query.Frequency)
	return &domain.SubscriptionList{
		Subscriptions: subscriptions,
		LastIndex:     int(resp.LastId),
	}, nil
}
