package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports/out"
	"weather-api/internal/core/service"
)

type ConfirmSubscriptionUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	tokenService     service.TokenService
	emailService     service.EmailService
}

func NewConfirmSubscriptionUseCase(
	subscriptionRepo out.SubscriptionRepository,
	tokenService service.TokenService,
	emailService service.EmailService,
) *ConfirmSubscriptionUseCase {
	return &ConfirmSubscriptionUseCase{
		subscriptionRepo: subscriptionRepo,
		tokenService:     tokenService,
		emailService:     emailService,
	}
}

func (uc *ConfirmSubscriptionUseCase) ConfirmSubscription(ctx context.Context, token string) error {
	log.Printf("Starting subscription confirmation for token: %s", token)

	if err := uc.tokenService.CheckTokenExists(ctx, token); err != nil {
		return err
	}

	subscription, err := uc.getSubscription(ctx, token)
	if err != nil {
		msg := fmt.Sprintf("unable to get subscription for token %s: %v", token, err)
		log.Print(msg)
		return errors.New(msg)
	}

	if err := uc.confirmSubscription(ctx, subscription); err != nil {
		msg := fmt.Sprintf("unable to confirm subscription for token %s: %v", token, err)
		log.Print(msg)
		return errors.New(msg)
	}

	log.Printf("Successfully confirmed subscription for token: %s", token)
	return nil
}

func (uc *ConfirmSubscriptionUseCase) getSubscription(ctx context.Context, token string) (domain.Subscription, error) {
	subscription, err := uc.subscriptionRepo.GetSubscriptionByToken(ctx, token)
	if err != nil {
		msg := fmt.Sprintf("unable to get subscription: %v", err)
		log.Print(msg)
		return domain.Subscription{}, errors.New(msg)
	}
	return subscription, nil
}

func (uc *ConfirmSubscriptionUseCase) confirmSubscription(ctx context.Context, subscription domain.Subscription) error {
	if subscription.IsConfirmed {
		return domain.ErrSubscriptionAlreadyConfirmed
	}

	subscription.IsConfirmed = true

	if err := uc.subscriptionRepo.UpdateSubscription(ctx, subscription); err != nil {
		msg := fmt.Sprintf("unable to update subscription: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}

	return nil
}
