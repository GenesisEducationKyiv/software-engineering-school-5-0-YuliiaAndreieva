package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/repository"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, opts repository.SubscribeOptions) (string, error)
	Confirm(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
	GetSubscriptionsByFrequency(ctx context.Context, frequency domain.Frequency) ([]domain.Subscription, error)
}

type SubscriptionServiceImpl struct {
	repo          repository.SubscriptionRepository
	weatherClient weather.Provider
	tokenSvc      TokenService
	cityRepo      repository.CityRepository
	emailService  EmailService
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	cityRepo repository.CityRepository,
	weatherClient weather.Provider,
	tokenSvc TokenService,
	emailService EmailService,
) *SubscriptionServiceImpl {
	return &SubscriptionServiceImpl{
		repo:          repo,
		cityRepo:      cityRepo,
		weatherClient: weatherClient,
		tokenSvc:      tokenSvc,
		emailService:  emailService,
	}
}

func (s *SubscriptionServiceImpl) Subscribe(ctx context.Context, opts repository.SubscribeOptions) (string, error) {
	cityEntity, err := s.cityRepo.GetByName(ctx, opts.City)
	if err != nil {
		if !errors.Is(err, domain.ErrCityNotFound) {
			msg := fmt.Sprintf("unable to get city %s: %v", opts.City, err)
			log.Print(msg)
			return "", errors.New(msg)
		}

		if err := s.weatherClient.CheckCityExists(ctx, opts.City); err != nil {
			if errors.Is(err, domain.ErrCityNotFound) {
				log.Printf("City %s not found in weather service", opts.City)
				return "", domain.ErrCityNotFound
			}
			msg := fmt.Sprintf("unable to check city existence for %s: %v", opts.City, err)
			log.Print(msg)
			return "", errors.New(msg)
		}

		cityEntity, err = s.cityRepo.Create(ctx, domain.City{Name: opts.City})
		if err != nil {
			msg := fmt.Sprintf("unable to create city %s: %v", opts.City, err)
			log.Print(msg)
			return "", errors.New(msg)
		}
	}

	exists, err := s.repo.IsSubscriptionExists(ctx, repository.IsSubscriptionExistsOptions{
		Email:     opts.Email,
		CityID:    cityEntity.ID,
		Frequency: opts.Frequency,
	})
	if err != nil {
		msg := fmt.Sprintf("unable to check subscription existence: %v", err)
		log.Print(msg)
		return "", errors.New(msg)
	}
	if exists {
		log.Printf("Email %s already subscribed to city %s with frequency %s", opts.Email, opts.City, opts.Frequency)
		return "", domain.ErrEmailAlreadySubscribed
	}

	token, err := s.tokenSvc.GenerateToken()
	if err != nil {
		msg := fmt.Sprintf("unable to generate token: %v", err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	sub := domain.Subscription{
		Email:       opts.Email,
		CityID:      cityEntity.ID,
		Frequency:   opts.Frequency,
		Token:       token,
		IsConfirmed: false,
	}
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		msg := fmt.Sprintf("unable to create subscription in repository: %v", err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	sub.City = &cityEntity

	if err := s.emailService.SendConfirmationEmail(&sub); err != nil {
		msg := fmt.Sprintf("unable to send confirmation email: %v", err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	log.Printf("Successfully created subscription")
	return token, nil
}

func (s *SubscriptionServiceImpl) Confirm(ctx context.Context, token string) error {
	log.Printf("Attempting to confirm subscription")

	if token == "" {
		log.Printf("Invalid token provided: empty token")
		return domain.ErrInvalidToken
	}

	exists, err := s.repo.IsTokenExists(ctx, token)
	if err != nil {
		msg := fmt.Sprintf("unable to check token existence: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}
	if !exists {
		log.Printf("Token not found: %s", token)
		return domain.ErrTokenNotFound
	}

	sub, err := s.repo.GetSubscriptionByToken(ctx, token)
	if err != nil {
		msg := fmt.Sprintf("unable to get subscription: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}
	sub.IsConfirmed = true
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		msg := fmt.Sprintf("unable to update subscription confirmation: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}

	log.Printf("Successfully confirmed subscription")
	return nil
}

func (s *SubscriptionServiceImpl) Unsubscribe(ctx context.Context, token string) error {
	log.Printf("Attempting to unsubscribe")

	if token == "" {
		log.Printf("Invalid token provided: empty token")
		return domain.ErrInvalidToken
	}

	exists, err := s.repo.IsTokenExists(ctx, token)
	if err != nil {
		msg := fmt.Sprintf("unable to check token existence: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}
	if !exists {
		log.Printf("Token not found: %s", token)
		return domain.ErrTokenNotFound
	}

	if err := s.repo.DeleteSubscription(ctx, token); err != nil {
		msg := fmt.Sprintf("unable to delete subscription: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}

	log.Printf("Successfully unsubscribed")
	return nil
}

func (s *SubscriptionServiceImpl) GetSubscriptionsByFrequency(ctx context.Context, frequency domain.Frequency) ([]domain.Subscription, error) {
	subscriptions, err := s.repo.GetSubscriptionsByFrequency(ctx, string(frequency))
	if err != nil {
		msg := fmt.Sprintf("unable to get subscriptions by frequency %s: %v", frequency, err)
		log.Print(msg)
		return nil, errors.New(msg)
	}
	log.Printf("Successfully retrieved %d subscriptions for frequency %s", len(subscriptions), frequency)
	return subscriptions, nil
}
