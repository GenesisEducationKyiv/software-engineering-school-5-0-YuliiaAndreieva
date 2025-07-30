package in

import (
	"context"
	"subscription/internal/core/domain"
)

type SubscribeUseCase interface {
	Subscribe(ctx context.Context, req domain.SubscriptionRequest) (*domain.SubscriptionResponse, error)
}

type ConfirmSubscriptionUseCase interface {
	ConfirmSubscription(ctx context.Context, token string) (*domain.ConfirmResponse, error)
}

type UnsubscribeUseCase interface {
	Unsubscribe(ctx context.Context, token string) (*domain.UnsubscribeResponse, error)
}

type ListByFrequencyUseCase interface {
	ListByFrequency(ctx context.Context, query domain.ListSubscriptionsQuery) (*domain.SubscriptionList, error)
}
