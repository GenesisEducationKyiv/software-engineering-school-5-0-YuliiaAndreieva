package usecase

import (
	"context"
	"fmt"
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

	if err := uc.validateToken(req.Token); err != nil {
		return uc.createErrorResponse(err.Error()), nil
	}

	if err := uc.validateTokenService(ctx, req.Token); err != nil {
		return uc.createErrorResponse("Invalid token"), nil
	}

	if err := uc.checkSubscriptionExists(ctx, req.Token); err != nil {
		return uc.createErrorResponse("Subscription not found"), err
	}

	if err := uc.deleteSubscription(ctx, req.Token); err != nil {
		return uc.createErrorResponse("Failed to unsubscribe"), err
	}

	uc.logger.Infof("Successfully unsubscribed for token: %s", req.Token)
	return uc.createSuccessResponse(), nil
}

func (uc *UnsubscribeUseCase) validateToken(token string) error {
	if token == "" {
		uc.logger.Errorf("Token is empty")
		return fmt.Errorf("token is required")
	}
	return nil
}

func (uc *UnsubscribeUseCase) validateTokenService(ctx context.Context, token string) error {
	isValid, err := uc.tokenService.ValidateToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to validate token %s: %v", token, err)
		return err
	}

	if !isValid {
		uc.logger.Errorf("Invalid token: %s", token)
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (uc *UnsubscribeUseCase) checkSubscriptionExists(ctx context.Context, token string) error {
	_, err := uc.subscriptionRepo.GetByToken(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to get subscription by token %s: %v", token, err)
		return err
	}
	return nil
}

func (uc *UnsubscribeUseCase) deleteSubscription(ctx context.Context, token string) error {
	err := uc.subscriptionRepo.Delete(ctx, token)
	if err != nil {
		uc.logger.Errorf("Failed to delete subscription for token %s: %v", token, err)
		return err
	}
	return nil
}

func (uc *UnsubscribeUseCase) createErrorResponse(message string) *domain.UnsubscribeResponse {
	return &domain.UnsubscribeResponse{
		Success: false,
		Message: message,
	}
}

func (uc *UnsubscribeUseCase) createSuccessResponse() *domain.UnsubscribeResponse {
	return &domain.UnsubscribeResponse{
		Success: true,
		Message: "Successfully unsubscribed",
	}
}
