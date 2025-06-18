package service

import (
	"context"
	"errors"
	"log"
	"weather-api/internal/adapter/email"
	"weather-api/internal/adapter/repository/postgres"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
	"weather-api/internal/util/emailutil"
)

type SubscriptionService struct {
	repo          postgres.SubscriptionRepository
	weatherSvc    WeatherService
	weatherClient weather.Provider
	emailSvc      email.EmailSender
	tokenSvc      TokenService
	cityRepo      postgres.CityRepository
}

func NewSubscriptionService(
	repo postgres.SubscriptionRepository,
	cityRepo postgres.CityRepository,
	weatherSvc WeatherService,
	weatherClient weather.Provider,
	emailSvc email.EmailSender,
	tokenSvc TokenService,
) *SubscriptionService {
	return &SubscriptionService{
		repo:          repo,
		cityRepo:      cityRepo,
		weatherSvc:    weatherSvc,
		weatherClient: weatherClient,
		emailSvc:      emailSvc,
		tokenSvc:      tokenSvc,
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, email, city string, frequency domain.Frequency) (string, error) {
	cityEntity, err := s.cityRepo.GetByName(ctx, city)
	if err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			if err := s.weatherClient.CheckCityExists(ctx, city); err != nil {
				if errors.Is(err, domain.ErrCityNotFound) {
					return "", domain.ErrCityNotFound
				}
				return "", err
			}
			cityEntity, err = s.cityRepo.Create(ctx, domain.City{Name: city})
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	exists, err := s.repo.IsSubscriptionExists(ctx, email, cityEntity.ID, frequency)
	if err != nil {
		return "", err
	}
	if exists {
		return "", domain.ErrEmailAlreadySubscribed
	}

	token, err := s.tokenSvc.GenerateToken()
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return "", err
	}

	sub := domain.Subscription{
		Email:       email,
		CityID:      cityEntity.ID,
		Frequency:   frequency,
		Token:       token,
		IsConfirmed: false,
	}
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		log.Printf("Failed to create subscription in repository: %v", err)
		return "", err
	}

	subject, htmlBody := emailutil.BuildConfirmationEmail(city, token)
	err = s.emailSvc.SendEmail(email, subject, htmlBody)
	if err != nil {
		log.Printf("Failed to send confirmation email: %v", err)
		return "", err
	}

	log.Printf("Successfully created subscription")
	return token, nil
}

func (s *SubscriptionService) Confirm(ctx context.Context, token string) error {
	log.Printf("Attempting to confirm subscription")

	if token == "" {
		return domain.ErrInvalidToken
	}

	exists, err := s.repo.IsTokenExists(ctx, token)
	if err != nil {
		log.Printf("Failed to check token existence: %v", err)
		return err
	}
	if !exists {
		return domain.ErrTokenNotFound
	}

	sub, err := s.repo.GetSubscriptionByToken(ctx, token)
	if err != nil {
		log.Printf("Failed to get subscription: %v", err)
		return err
	}
	sub.IsConfirmed = true
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		log.Printf("Failed to update subscription confirmation: %v", err)
		return err
	}

	log.Printf("Successfully confirmed subscription")
	return nil
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	log.Printf("Attempting to unsubscribe")

	if token == "" {
		return domain.ErrInvalidToken
	}

	exists, err := s.repo.IsTokenExists(ctx, token)
	if err != nil {
		log.Printf("Failed to check token existence: %v", err)
		return err
	}
	if !exists {
		return domain.ErrTokenNotFound
	}

	if err := s.repo.DeleteSubscription(ctx, token); err != nil {
		log.Printf("Failed to delete subscription: %v", err)
		return err
	}

	log.Printf("Successfully unsubscribed")
	return nil
}

func (s *SubscriptionService) PrepareUpdates(ctx context.Context, frequency domain.Frequency) ([]domain.WeatherUpdate, error) {
	subs, err := s.repo.GetSubscriptionsByFrequency(ctx, string(frequency))
	if err != nil {
		return nil, err
	}

	citySubscriptions := make(map[string][]domain.Subscription)
	for _, sub := range subs {
		if !sub.IsConfirmed {
			continue
		}
		citySubscriptions[sub.City.Name] = append(citySubscriptions[sub.City.Name], sub)
	}

	var updates []domain.WeatherUpdate
	for cityName, citySubs := range citySubscriptions {
		weather, err := s.weatherSvc.GetWeather(ctx, cityName)
		if err != nil {
			log.Printf("Failed to get weather for city %s: %v", cityName, err)
			continue
		}

		for _, sub := range citySubs {
			updates = append(updates, domain.WeatherUpdate{
				Subscription: sub,
				Weather:      weather,
			})
		}
	}

	return updates, nil
}
