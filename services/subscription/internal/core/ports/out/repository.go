package out

import (
	"context"
	"subscription/internal/core/domain"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription domain.Subscription) error
	GetByToken(ctx context.Context, token string) (*domain.Subscription, error)
	Update(ctx context.Context, subscription domain.Subscription) error
	Delete(ctx context.Context, token string) error
	ListByFrequency(ctx context.Context, frequency domain.Frequency, lastID int, pageSize int) (*domain.SubscriptionList, error)
}
