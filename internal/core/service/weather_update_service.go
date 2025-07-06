package service

import (
	"context"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
)

type WeatherUpdateService interface {
	PrepareUpdates(ctx context.Context, frequency domain.Frequency) ([]domain.WeatherUpdate, error)
}

type WeatherUpdateServiceImpl struct {
	subscriptionService SubscriptionService
	weatherService      WeatherService
}

func NewWeatherUpdateService(
	subscriptionService SubscriptionService,
	weatherService WeatherService,
) WeatherUpdateService {
	return &WeatherUpdateServiceImpl{
		subscriptionService: subscriptionService,
		weatherService:      weatherService,
	}
}

func (s *WeatherUpdateServiceImpl) PrepareUpdates(ctx context.Context, frequency domain.Frequency) ([]domain.WeatherUpdate, error) {
	subs, err := s.subscriptionService.GetSubscriptionsByFrequency(ctx, frequency)
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
		weather, err := s.weatherService.GetWeather(ctx, cityName)
		if err != nil {
			msg := fmt.Sprintf("unable to get weather for city %s: %v", cityName, err)
			log.Print(msg)
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
