package out

import (
	"context"
	"weather-broadcast/internal/core/domain"
)

//go:generate mockery --name WeatherClient
type WeatherClient interface {
	GetWeatherByCity(ctx context.Context, city string) (*domain.Weather, error)
}
