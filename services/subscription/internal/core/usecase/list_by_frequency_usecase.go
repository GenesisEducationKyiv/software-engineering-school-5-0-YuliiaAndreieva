package usecase

import (
	"context"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"
)

type ListByFrequencyUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	logger           out.Logger
}

func NewListByFrequencyUseCase(subscriptionRepo out.SubscriptionRepository, logger out.Logger) in.ListByFrequencyUseCase {
	return &ListByFrequencyUseCase{
		subscriptionRepo: subscriptionRepo,
		logger:           logger,
	}
}

func (uc *ListByFrequencyUseCase) ListByFrequency(ctx context.Context, query domain.ListSubscriptionsQuery) (*domain.SubscriptionList, error) {
	uc.logger.Infof("Listing subscriptions by frequency: %s, lastID: %d, pageSize: %d", query.Frequency, query.LastID, query.PageSize)

	subscriptions, err := uc.subscriptionRepo.ListByFrequency(ctx, domain.Frequency(query.Frequency), query.LastID, query.PageSize)
	if err != nil {
		uc.logger.Errorf("Failed to list subscriptions: %v", err)
		return nil, err
	}

	uc.logger.Infof("Found %d subscriptions", len(subscriptions.Subscriptions))
	return subscriptions, nil
}
