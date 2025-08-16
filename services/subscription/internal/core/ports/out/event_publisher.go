package out

import (
	"context"
	"subscription/internal/core/domain"
)

type EventPublisher interface {
	PublishSubscriptionCreated(ctx context.Context, subscription domain.Subscription) error
}
