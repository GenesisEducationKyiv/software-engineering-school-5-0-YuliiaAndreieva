package port

import (
	"context"
	"weather-api/internal/core/domain"
)

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub domain.Subscription) error
	GetSubscriptionByToken(ctx context.Context, token string) (domain.Subscription, error)
	UpdateSubscription(ctx context.Context, sub domain.Subscription) error
	DeleteSubscription(ctx context.Context, token string) error
	GetSubscriptionsByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error)
	IsEmailSubscribed(ctx context.Context, email string) (bool, error)
	IsTokenExists(ctx context.Context, token string) (bool, error)
}
