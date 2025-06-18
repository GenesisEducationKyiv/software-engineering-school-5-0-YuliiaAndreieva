package service

import (
	"context"
	"log"
	"weather-api/internal/core/domain"
)

type WeatherUpdateService struct {
	subscriptionService *SubscriptionService
	weatherService      WeatherService
}

func NewWeatherUpdateService(
	subscriptionService *SubscriptionService,
	weatherService WeatherService,
) *WeatherUpdateService {
	return &WeatherUpdateService{
		subscriptionService: subscriptionService,
		weatherService:      weatherService,
	}
}

func (s *WeatherUpdateService) PrepareUpdates(ctx context.Context, frequency domain.Frequency) ([]domain.WeatherUpdate, error) {
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
