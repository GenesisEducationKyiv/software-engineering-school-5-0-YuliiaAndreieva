package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"weather-api/internal/core/ports/out"
	"weather-api/internal/core/service"
)

type UnsubscribeUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	tokenService     service.TokenService
}

func NewUnsubscribeUseCase(
	subscriptionRepo out.SubscriptionRepository,
	tokenService service.TokenService,
) *UnsubscribeUseCase {
	return &UnsubscribeUseCase{
		subscriptionRepo: subscriptionRepo,
		tokenService:     tokenService,
	}
}

func (uc *UnsubscribeUseCase) Unsubscribe(ctx context.Context, token string) error {
	log.Printf("Starting unsubscribe process for token: %s", token)

	if err := uc.tokenService.CheckTokenExists(ctx, token); err != nil {
		msg := fmt.Sprintf("unable to validate token %s: %v", token, err)
		log.Print(msg)
		return errors.New(msg)
	}

	if err := uc.subscriptionRepo.DeleteSubscription(ctx, token); err != nil {
		msg := fmt.Sprintf("unable to delete subscription for token %s: %v", token, err)
		log.Print(msg)
		return errors.New(msg)
	}

	log.Printf("Successfully unsubscribed for token: %s", token)
	return nil
}
