package ports

import (
	"context"
	"weather-api/internal/core/domain"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, opts SubscribeOptions) (string, error)
	Confirm(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
	GetSubscriptionsByFrequency(ctx context.Context, frequency domain.Frequency) ([]domain.Subscription, error)
}

type WeatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
}
