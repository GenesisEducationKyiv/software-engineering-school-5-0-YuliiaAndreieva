package usecase

import (
	"context"
	"subscription-service/internal/core/domain"
	"subscription-service/internal/core/ports/in"
	"subscription-service/internal/core/ports/out"
)

type SubscribeUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	tokenService     out.TokenService
	emailService     out.EmailService
	logger           out.Logger
}

func NewSubscribeUseCase(
	subscriptionRepo out.SubscriptionRepository,
	tokenService out.TokenService,
	emailService out.EmailService,
	logger out.Logger,
) in.SubscribeUseCase {
	return &SubscribeUseCase{
		subscriptionRepo: subscriptionRepo,
		tokenService:     tokenService,
		emailService:     emailService,
		logger:           logger,
	}
}

func (uc *SubscribeUseCase) Subscribe(ctx context.Context, req domain.SubscriptionRequest) (*domain.SubscriptionResponse, error) {
	uc.logger.Infof("Starting subscription process for email: %s, city: %s", req.Email, req.City)

	exists, err := uc.subscriptionRepo.IsSubscriptionExists(ctx, req.Email, req.City)
	if err != nil {
		uc.logger.Errorf("Failed to check subscription existence: %v", err)
		return nil, err
	}

	if exists {
		uc.logger.Warnf("Email %s already subscribed to city %s", req.Email, req.City)
		return &domain.SubscriptionResponse{
			Success: false,
			Message: "Email already subscribed to this city",
		}, nil
	}

	token, err := uc.tokenService.GenerateToken(ctx, req.Email, "24h")
	if err != nil {
		uc.logger.Errorf("Failed to generate token: %v", err)
		return nil, err
	}

	subscription := domain.Subscription{
		Email:       req.Email,
		City:        req.City,
		Frequency:   req.Frequency,
		Token:       token,
		IsConfirmed: false,
	}

	if err := uc.subscriptionRepo.CreateSubscription(ctx, subscription); err != nil {
		uc.logger.Errorf("Failed to create subscription: %v", err)
		return nil, err
	}

	if err := uc.emailService.SendConfirmationEmail(ctx, req.Email, req.City, token); err != nil {
		uc.logger.Errorf("Failed to send confirmation email: %v", err)
		return &domain.SubscriptionResponse{
			Success: true,
			Message: "Subscription created but confirmation email failed to send",
			Token:   token,
		}, nil
	}

	uc.logger.Infof("Successfully created subscription for email: %s, city: %s", req.Email, req.City)
	return &domain.SubscriptionResponse{
		Success: true,
		Message: "Subscription successful. Confirmation email sent.",
		Token:   token,
	}, nil
}
