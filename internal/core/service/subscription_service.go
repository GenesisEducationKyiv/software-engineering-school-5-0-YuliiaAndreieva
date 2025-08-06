package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports/out"
)

type TokenService interface {
	GenerateToken() (string, error)
	CheckTokenExists(ctx context.Context, token string) error
}
type SubscriptionServiceImpl struct {
	repo          out.SubscriptionRepository
	weatherClient out.WeatherProvider
	tokenSvc      TokenService
	cityRepo      out.CityRepository
	emailService  EmailService
}

func NewSubscriptionService(
	repo out.SubscriptionRepository,
	cityRepo out.CityRepository,
	weatherClient out.WeatherProvider,
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

func (s *SubscriptionServiceImpl) CreateSubscription(ctx context.Context, email string, cityID int64, frequency domain.Frequency) (string, error) {
	token, err := s.tokenSvc.GenerateToken()
	if err != nil {
		msg := fmt.Sprintf("unable to generate token: %v", err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	subscription := domain.Subscription{
		Email:       email,
		CityID:      cityID,
		Frequency:   frequency,
		Token:       token,
		IsConfirmed: false,
	}

	if err := s.repo.CreateSubscription(ctx, subscription); err != nil {
		msg := fmt.Sprintf("unable to create subscription in repository: %v", err)
		log.Print(msg)
		return "", errors.New(msg)
	}

	return token, nil
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
