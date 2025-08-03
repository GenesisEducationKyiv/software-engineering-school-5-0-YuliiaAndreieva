package usecase

import (
	"context"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"
)

type ConfirmSubscriptionUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	tokenService     out.TokenService
	logger           out.Logger
}

func NewConfirmSubscriptionUseCase(
	subscriptionRepo out.SubscriptionRepository,
	tokenService out.TokenService,
	logger out.Logger,
) in.ConfirmSubscriptionUseCase {
	return &ConfirmSubscriptionUseCase{
		subscriptionRepo: subscriptionRepo,
		tokenService:     tokenService,
		logger:           logger,
	}
}

func (uc *ConfirmSubscriptionUseCase) ConfirmSubscription(ctx context.Context, token string) (*domain.ConfirmResponse, error) {
	uc.logger.Infof("Starting subscription confirmation for token: %s", token)

	valid, err := uc.tokenService.ValidateToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to validate token: %v", err)
		return &domain.ConfirmResponse{
			Success: false,
			Message: "Token validation failed",
		}, nil
	}

	if !valid {
		uc.logger.Errorf("Invalid token: %s", token)
		return &domain.ConfirmResponse{
			Success: false,
			Message: "Invalid token",
		}, nil
	}

	subscription, err := uc.subscriptionRepo.GetByToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to get subscription by token: %v", err)
		return &domain.ConfirmResponse{
			Success: false,
			Message: "Token not found",
		}, nil
	}

	if subscription.IsConfirmed {
		uc.logger.Warnf("Subscription already confirmed for token: %s", token)
		return &domain.ConfirmResponse{
			Success: false,
			Message: "Subscription already confirmed",
		}, nil
	}

	subscription.IsConfirmed = true
	if err := uc.subscriptionRepo.Update(ctx, *subscription); err != nil {
		uc.logger.Errorf("Failed to update subscription: %v", err)
		return nil, err
	}

	uc.logger.Infof("Successfully confirmed subscription for token: %s", token)
	return &domain.ConfirmResponse{
		Success: true,
		Message: "Subscription confirmed successfully",
	}, nil
}
