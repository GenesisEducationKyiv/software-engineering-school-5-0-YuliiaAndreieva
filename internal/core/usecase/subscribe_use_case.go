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

type CityService interface {
	EnsureCityExists(ctx context.Context, cityName string) (domain.City, error)
}

type SubscribeUseCase struct {
	subscriptionRepo out.SubscriptionRepository
	subscriptionSvc  out.SubscriptionService
	cityService      CityService
	tokenService     service.TokenService
	emailService     service.EmailService
}

func NewSubscribeUseCase(
	subscriptionRepo out.SubscriptionRepository,
	subscriptionSvc out.SubscriptionService,
	cityService CityService,
	tokenService service.TokenService,
	emailService service.EmailService,
) *SubscribeUseCase {
	return &SubscribeUseCase{
		subscriptionRepo: subscriptionRepo,
		subscriptionSvc:  subscriptionSvc,
		cityService:      cityService,
		tokenService:     tokenService,
		emailService:     emailService,
	}
}

func (uc *SubscribeUseCase) Subscribe(ctx context.Context, opts out.SubscribeOptions) (string, error) {
	log.Printf("Starting subscription process for email: %s, city: %s, frequency: %s", opts.Email, opts.City, opts.Frequency)

	city, err := uc.ensureCityExists(ctx, opts.City)
	if err != nil {
		msg := fmt.Sprintf("unable to check city existence for %s: %v", opts.City, err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	if err := uc.checkExistingSubscription(ctx, opts, city.ID); err != nil {
		msg := fmt.Sprintf("unable to check subscription existence for %s: %v", opts.Email, err)
		log.Print(msg)
		return "", err
	}

	token, err := uc.createSubscription(ctx, opts, city)
	if err != nil {
		msg := fmt.Sprintf("unable to create subscription for %s: %v", opts.Email, err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	if err := uc.sendConfirmationEmail(city, token, opts.Email); err != nil {
		msg := fmt.Sprintf("unable to send confirmation email for %s: %v", opts.Email, err)
		log.Print(msg)
	}

	log.Printf("Successfully created subscription for email: %s, city: %s", opts.Email, opts.City)
	return token, nil
}

func (uc *SubscribeUseCase) ensureCityExists(ctx context.Context, cityName string) (domain.City, error) {
	city, err := uc.cityService.EnsureCityExists(ctx, cityName)
	if err != nil {
		return domain.City{}, err
	}

	return city, nil
}

func (uc *SubscribeUseCase) checkExistingSubscription(ctx context.Context, opts out.SubscribeOptions, cityID int64) error {
	exists, err := uc.subscriptionRepo.IsSubscriptionExists(ctx, out.IsSubscriptionExistsOptions{
		Email:     opts.Email,
		CityID:    cityID,
		Frequency: opts.Frequency,
	})
	if err != nil {
		msg := fmt.Sprintf("unable to check subscription existence: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}

	if exists {
		log.Printf("Email %s already subscribed to city %s with frequency %s", opts.Email, opts.City, opts.Frequency)
		return domain.ErrEmailAlreadySubscribed
	}

	return nil
}

func (uc *SubscribeUseCase) createSubscription(ctx context.Context, opts out.SubscribeOptions, city domain.City) (string, error) {
	token, err := uc.subscriptionSvc.CreateSubscription(ctx, opts.Email, city.ID, opts.Frequency)
	if err != nil {
		msg := fmt.Sprintf("unable to create subscription: %v", err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	return token, nil
}

func (uc *SubscribeUseCase) sendConfirmationEmail(city domain.City, token string, email string) error {
	subscription := &domain.Subscription{
		Email: email,
		City:  &city,
		Token: token,
	}

	return uc.emailService.SendConfirmationEmail(subscription)
}
