package out

import (
	"context"
	"weather-api/internal/core/domain"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, email string, cityID int64, frequency domain.Frequency) (string, error)
	GetSubscriptionsByFrequency(ctx context.Context, frequency domain.Frequency) ([]domain.Subscription, error)
}

type WeatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
}

type CityService interface {
	EnsureCityExists(ctx context.Context, cityName string) (domain.City, error)
}
