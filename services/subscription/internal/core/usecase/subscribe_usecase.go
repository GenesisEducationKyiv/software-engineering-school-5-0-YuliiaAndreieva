package usecase

import (
	"context"
	"fmt"
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
	uc.logger.Infof("Starting subscription process for email: %s, city: %s, frequency: %s", req.Email, req.City, req.Frequency)

	uc.logger.Debugf("Checking if subscription already exists for email: %s and city: %s", req.Email, req.City)
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

	uc.logger.Debugf("Generating token for email: %s", req.Email)
	token, err := uc.tokenService.GenerateToken(ctx, req.Email, "24h")
	if err != nil {
		uc.logger.Errorf("Failed to generate token: %v", err)
		return nil, err
	}
	uc.logger.Debugf("Token generated successfully for email: %s", req.Email)

	subscription := domain.Subscription{
		Email:       req.Email,
		City:        req.City,
		Frequency:   req.Frequency,
		Token:       token,
		IsConfirmed: false,
	}

	uc.logger.Debugf("Creating subscription in database for email: %s, city: %s", req.Email, req.City)
	if err := uc.subscriptionRepo.CreateSubscription(ctx, subscription); err != nil {
		uc.logger.Errorf("Failed to create subscription: %v", err)
		return nil, err
	}
	uc.logger.Infof("Subscription created successfully in database for email: %s, city: %s", req.Email, req.City)

	uc.logger.Debugf("Sending confirmation email to: %s for city: %s", req.Email, req.City)

	confirmationReq := domain.ConfirmationEmailRequest{
		To:               req.Email,
		Subject:          "Confirm your weather subscription",
		City:             req.City,
		ConfirmationLink: fmt.Sprintf("%s/confirm/%s", "http://localhost:8082", token),
	}

	_, err = uc.emailService.SendConfirmationEmail(ctx, confirmationReq)
	if err != nil {
		uc.logger.Errorf("Failed to send confirmation email: %v", err)
		uc.logger.Warnf("Subscription created but email delivery failed for email: %s", req.Email)
		return &domain.SubscriptionResponse{
			Success: true,
			Message: "Subscription created but confirmation email failed to send",
			Token:   token,
		}, nil
	}

	uc.logger.Infof("Successfully created subscription for email: %s, city: %s", req.Email, req.City)
	uc.logger.Infof("Confirmation email sent successfully to: %s", req.Email)
	return &domain.SubscriptionResponse{
		Success: true,
		Message: "Subscription successful. Confirmation email sent.",
		Token:   token,
	}, nil
}
