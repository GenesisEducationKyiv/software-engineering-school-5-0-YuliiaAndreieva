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
	emailService     out.EmailService
	logger           out.Logger
	config           *config.Config
}

func NewSubscribeUseCase(
	subscriptionRepo out.SubscriptionRepository,
	tokenService out.TokenService,
	emailService out.EmailService,
	logger out.Logger,
	config *config.Config,
) in.SubscribeUseCase {
	return &SubscribeUseCase{
		subscriptionRepo: subscriptionRepo,
		tokenService:     tokenService,
		emailService:     emailService,
		logger:           logger,
		config:           config,
	}
}

func (uc *SubscribeUseCase) Subscribe(ctx context.Context, req domain.SubscriptionRequest) (*domain.SubscriptionResponse, error) {
	uc.logger.Infof("Starting subscription process for email: %s, city: %s, frequency: %s", req.Email, req.City, req.Frequency)

	token, err := uc.generateToken(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	subscription := uc.createSubscription(req, token)

	if err := uc.saveSubscription(ctx, subscription); err != nil {
		return uc.handleSubscriptionSaveError(err)
	}

	confirmationReq := uc.createConfirmationEmailRequest(req, token)
	emailErr := uc.sendConfirmationEmail(ctx, confirmationReq)

	return uc.createResponse(token, emailErr)
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

func (uc *SubscribeUseCase) createConfirmationEmailRequest(req domain.SubscriptionRequest, token string) domain.ConfirmationEmailRequest {
	return domain.ConfirmationEmailRequest{
		To:               req.Email,
		Subject:          "Confirm your weather subscription",
		City:             req.City,
		ConfirmationLink: fmt.Sprintf("%s/confirm/%s", "http://localhost:8082", token),
	}
}

func (uc *SubscribeUseCase) sendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) error {
	uc.logger.Debugf("Sending confirmation email to: %s for city: %s", req.To, req.City)
	err := uc.emailService.SendConfirmationEmail(ctx, req)
	if err != nil {
		uc.logger.Errorf("Failed to send confirmation email: %v", err)
		uc.logger.Warnf("Subscription created but email delivery failed for email: %s", req.To)
		return err
	}
	uc.logger.Infof("Confirmation email sent successfully to: %s", req.To)
	return nil
}

func (uc *SubscribeUseCase) createResponse(token string, emailErr error) (*domain.SubscriptionResponse, error) {
	if emailErr != nil {
		return &domain.SubscriptionResponse{
			Success: true,
			Message: "Subscription created but confirmation email failed to send",
			Token:   token,
		}, nil
	}

	return &domain.SubscriptionResponse{
		Success: true,
		Message: "Subscription successful. Confirmation email sent.",
		Token:   token,
	}, nil
}
