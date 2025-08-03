package usecase

import (
	"context"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"
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

func (uc *UnsubscribeUseCase) Unsubscribe(ctx context.Context, req domain.UnsubscribeRequest) (*domain.UnsubscribeResponse, error) {
	uc.logger.Infof("Starting subscription unsubscription for token: %s", req.Token)

	if req.Token == "" {
		uc.logger.Errorf("Token is empty")
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Token is required",
		}, nil
	}

	isValid, err := uc.tokenService.ValidateToken(ctx, req.Token)
	if err != nil {
		uc.logger.Errorf("Failed to validate token %s: %v", req.Token, err)
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Failed to validate token",
		}, err
	}

	if !isValid {
		uc.logger.Errorf("Invalid token: %s", req.Token)
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Invalid token",
		}, nil
	}

	_, err = uc.subscriptionRepo.GetByToken(ctx, req.Token)
	if err != nil {
		uc.logger.Errorf("Failed to get subscription by token %s: %v", req.Token, err)
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Subscription not found",
		}, err
	}

	err = uc.subscriptionRepo.Delete(ctx, req.Token)
	if err != nil {
		uc.logger.Errorf("Failed to delete subscription for token %s: %v", req.Token, err)
		return &domain.UnsubscribeResponse{
			Success: false,
			Message: "Failed to unsubscribe",
		}, err
	}

	uc.logger.Infof("Successfully unsubscribed for token: %s", req.Token)
	return &domain.UnsubscribeResponse{
		Success: true,
		Message: "Successfully unsubscribed",
	}, nil
}
