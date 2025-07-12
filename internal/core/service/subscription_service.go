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

func (s *SubscriptionServiceImpl) Subscribe(ctx context.Context, opts out.SubscribeOptions) (string, error) {
	cityEntity, err := s.ensureCityExists(ctx, opts.City)
	if err != nil {
		return "", err
	}

	if err := s.checkSubscriptionExists(ctx, opts, cityEntity.ID); err != nil {
		return "", err
	}

	token, err := s.createSubscription(ctx, opts, cityEntity)
	if err != nil {
		return "", err
	}

	log.Printf("Successfully created subscription")
	return token, nil
}

func (s *SubscriptionServiceImpl) ensureCityExists(ctx context.Context, cityName string) (domain.City, error) {
	city, err := s.cityRepo.GetByName(ctx, cityName)

	if err == nil {
		return city, nil
	}

	if !errors.Is(err, domain.ErrCityNotFound) {
		msg := fmt.Sprintf("unable to get city %s: %v", cityName, err)
		log.Print(msg)
		return domain.City{}, errors.New(msg)
	}

	if err := s.weatherClient.CheckCityExists(ctx, cityName); err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			log.Printf("City %s not found in weather service", cityName)
			return domain.City{}, domain.ErrCityNotFound
		}
		msg := fmt.Sprintf("unable to check city existence for %s: %v", cityName, err)
		log.Print(msg)
		return domain.City{}, errors.New(msg)
	}

	city, err = s.cityRepo.Create(ctx, domain.City{Name: cityName})
	if err != nil {
		msg := fmt.Sprintf("unable to create city %s: %v", cityName, err)
		log.Print(msg)
		return domain.City{}, errors.New(msg)
	}

	return city, nil
}

func (s *SubscriptionServiceImpl) checkSubscriptionExists(ctx context.Context, opts out.SubscribeOptions, cityID int64) error {
	exists, err := s.repo.IsSubscriptionExists(ctx, out.IsSubscriptionExistsOptions{
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

func (s *SubscriptionServiceImpl) createSubscription(ctx context.Context, opts out.SubscribeOptions, cityEntity domain.City) (string, error) {
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

	return token, nil
}

func (s *SubscriptionServiceImpl) Confirm(ctx context.Context, token string) error {
	log.Printf("Attempting to confirm subscription")

	if err := s.tokenSvc.CheckTokenExists(ctx, token); err != nil {
		return err
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

	if err := s.tokenSvc.CheckTokenExists(ctx, token); err != nil {
		return err
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
