package service

import (
	"context"
	"errors"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/port"
	"weather-api/internal/util"
)

type SubscriptionService struct {
	repo       port.SubscriptionRepository
	weatherSvc port.WeatherService
	emailSvc   port.EmailService
	tokenSvc   port.TokenService
}

func NewSubscriptionService(repo port.SubscriptionRepository, weatherSvc port.WeatherService, emailSvc port.EmailService, tokenSvc port.TokenService) *SubscriptionService {
	return &SubscriptionService{
		repo:       repo,
		weatherSvc: weatherSvc,
		emailSvc:   emailSvc,
		tokenSvc:   tokenSvc,
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, email string, city string, frequency domain.Frequency) (string, error) {
	log.Printf("Attempting to create subscription for city: %s, frequency: %s", city, frequency)

	isSubscribed, err := s.repo.IsEmailSubscribed(ctx, email)
	if err != nil {
		log.Printf("Failed to check email subscription: %v", err)
		return "", err
	}
	if isSubscribed {
		return "", domain.ErrEmailAlreadySubscribed
	}

	_, err = s.weatherSvc.GetWeather(city)
	if err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			log.Printf("City not found: %s", city)
			return "", domain.ErrCityNotFound
		}
		log.Printf("Failed to validate city: %v", err)
		return "", err
	}

	token, err := s.tokenSvc.GenerateToken()
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return "", err
	}

	sub := domain.Subscription{
		Email:       email,
		City:        city,
		Frequency:   frequency,
		Token:       token,
		IsConfirmed: false,
	}
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		log.Printf("Failed to create subscription in repository: %v", err)
		return "", err
	}

	subject, htmlBody := util.BuildConfirmationEmail(city, token)
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
	//TODO: validate and generate the token better in future

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
