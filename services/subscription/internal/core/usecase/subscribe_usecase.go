package usecase

import (
	"context"
	"fmt"
	"subscription/internal/config"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"
)

type SubscribeUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	tokenService     out.TokenService
	eventPublisher   out.EventPublisher
	logger           out.Logger
	config           *config.Config
}

func NewSubscribeUseCase(
	subscriptionRepo out.SubscriptionRepository,
	tokenService out.TokenService,
	eventPublisher out.EventPublisher,
	logger out.Logger,
	config *config.Config,
) in.SubscribeUseCase {
	return &SubscribeUseCase{
		subscriptionRepo: subscriptionRepo,
		tokenService:     tokenService,
		eventPublisher:   eventPublisher,
		logger:           logger,
		config:           config,
	}
}

func (uc *SubscribeUseCase) Subscribe(ctx context.Context, req domain.SubscriptionRequest) (*domain.SubscriptionResponse, error) {
	uc.logger.Infof("Processing subscription request for email: %s, city: %s", req.Email, req.City)

	token, err := uc.generateToken(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	subscription := uc.createSubscription(req, token)

	if err := uc.saveSubscription(ctx, subscription); err != nil {
		return uc.handleSubscriptionSaveError(err)
	}

	if err := uc.publishSubscriptionCreated(ctx, subscription); err != nil {
		uc.logger.Errorf("Failed to publish subscription created event: %v", err)
	}

	return uc.createResponse(token)
}

func (uc *SubscribeUseCase) generateToken(ctx context.Context, email string) (string, error) {
	uc.logger.Debugf("Generating token for email: %s with expiration: %s", email, uc.config.Token.Expiration)
	token, err := uc.tokenService.GenerateToken(ctx, email, uc.config.Token.Expiration)
	if err != nil {
		uc.logger.Errorf("Failed to generate token: %v", err)
		return "", err
	}
	uc.logger.Debugf("Token generated successfully for email: %s", email)
	return token, nil
}

func (uc *SubscribeUseCase) createSubscription(req domain.SubscriptionRequest, token string) domain.Subscription {
	return domain.Subscription{
		Email:       req.Email,
		City:        req.City,
		Frequency:   req.Frequency,
		Token:       token,
		IsConfirmed: false,
	}
}

func (uc *SubscribeUseCase) saveSubscription(ctx context.Context, subscription domain.Subscription) error {
	uc.logger.Debugf("Creating subscription in database for email: %s, city: %s", subscription.Email, subscription.City)
	if err := uc.subscriptionRepo.Create(ctx, subscription); err != nil {
		uc.logger.Errorf("Failed to create subscription: %v", err)
		return err
	}
	uc.logger.Infof("Subscription created successfully in database for email: %s, city: %s", subscription.Email, subscription.City)
	return nil
}

func (uc *SubscribeUseCase) handleSubscriptionSaveError(err error) (*domain.SubscriptionResponse, error) {
	if err == domain.ErrDuplicateSubscription {
		return &domain.SubscriptionResponse{
			Success: false,
			Message: "Subscription already exists for this email, city and frequency",
		}, nil
	}
	return nil, err
}

func (uc *SubscribeUseCase) publishSubscriptionCreated(ctx context.Context, subscription domain.Subscription) error {
	if err := uc.eventPublisher.PublishSubscriptionCreated(ctx, subscription); err != nil {
		return fmt.Errorf("failed to publish subscription created event: %w", err)
	}
	return nil
}

func (uc *SubscribeUseCase) createResponse(token string) (*domain.SubscriptionResponse, error) {
	return &domain.SubscriptionResponse{
		Success: true,
		Message: "Subscription successful. Confirmation email will be sent shortly.",
		Token:   token,
	}, nil
}
