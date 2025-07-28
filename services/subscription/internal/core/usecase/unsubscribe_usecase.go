package usecase

import (
	"context"
	"subscription-service/internal/core/domain"
	"subscription-service/internal/core/ports/in"
	"subscription-service/internal/core/ports/out"
)

type UnsubscribeUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	tokenService     out.TokenService
	logger           out.Logger
}

func NewUnsubscribeUseCase(
	subscriptionRepo out.SubscriptionRepository,
	tokenService out.TokenService,
	logger out.Logger,
) in.UnsubscribeUseCase {
	return &UnsubscribeUseCase{
		subscriptionRepo: subscriptionRepo,
		tokenService:     tokenService,
		logger:           logger,
	}
}

func (uc *UnsubscribeUseCase) Unsubscribe(ctx context.Context, token string) (*domain.UnsubscribeResponse, error) {
	uc.logger.Infof("Starting unsubscribe process for token: %s", token)

	valid, err := uc.tokenService.ValidateToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to validate token: %v", err)
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Token validation failed",
		}, nil
	}

	if !valid {
		uc.logger.Errorf("Invalid token: %s", token)
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Invalid token",
		}, nil
	}

	_, err = uc.subscriptionRepo.GetSubscriptionByToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to get subscription by token: %v", err)
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Token not found",
		}, nil
	}

	if err := uc.subscriptionRepo.DeleteSubscription(ctx, token); err != nil {
		uc.logger.Errorf("Failed to delete subscription: %v", err)
		return nil, err
	}

	uc.logger.Infof("Successfully unsubscribed for token: %s", token)
	return &domain.UnsubscribeResponse{
		Success: true,
		Message: "Successfully unsubscribed",
	}, nil
}
