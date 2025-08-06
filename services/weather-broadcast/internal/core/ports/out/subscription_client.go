package out

import (
	"context"
	"weather-broadcast/internal/core/domain"
)

//go:generate mockery --name SubscriptionClient
type SubscriptionClient interface {
	ListByFrequency(ctx context.Context, query domain.ListSubscriptionsQuery) (*domain.SubscriptionList, error)
}
