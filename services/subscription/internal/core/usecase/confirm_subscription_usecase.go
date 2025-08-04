package usecase

import (
	"context"
	"fmt"
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

	if err := uc.validateToken(ctx, token); err != nil {
		return uc.createErrorResponse(err.Error()), nil
	}

	subscription, err := uc.getSubscriptionByToken(ctx, token)
	if err != nil {
		return uc.createErrorResponse("Token not found"), nil
	}

	if subscription.IsConfirmed {
		uc.logger.Warnf("Subscription already confirmed for token: %s", token)
		return uc.createErrorResponse("Subscription already confirmed"), nil
	}

	if err := uc.confirmSubscription(ctx, subscription); err != nil {
		return nil, err
	}

	uc.logger.Infof("Successfully confirmed subscription for token: %s", token)
	return uc.createSuccessResponse(), nil
}

func (uc *ConfirmSubscriptionUseCase) validateToken(ctx context.Context, token string) error {
	valid, err := uc.tokenService.ValidateToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to validate token: %v", err)
		return fmt.Errorf("Token validation failed")
	}

	if !valid {
		uc.logger.Errorf("Invalid token: %s", token)
		return fmt.Errorf("Invalid token")
	}

	return nil
}

func (uc *ConfirmSubscriptionUseCase) getSubscriptionByToken(ctx context.Context, token string) (*domain.Subscription, error) {
	subscription, err := uc.subscriptionRepo.GetByToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to get subscription by token: %v", err)
		return nil, err
	}
	return subscription, nil
}

func (uc *ConfirmSubscriptionUseCase) confirmSubscription(ctx context.Context, subscription *domain.Subscription) error {
	subscription.IsConfirmed = true
	if err := uc.subscriptionRepo.Update(ctx, *subscription); err != nil {
		uc.logger.Errorf("Failed to update subscription: %v", err)
		return err
	}
	return nil
}

func (uc *ConfirmSubscriptionUseCase) createErrorResponse(message string) *domain.ConfirmResponse {
	return &domain.ConfirmResponse{
		Success: false,
		Message: message,
	}
}

func (uc *ConfirmSubscriptionUseCase) createSuccessResponse() *domain.ConfirmResponse {
	return &domain.ConfirmResponse{
		Success: true,
		Message: "Subscription confirmed successfully",
	}
}
