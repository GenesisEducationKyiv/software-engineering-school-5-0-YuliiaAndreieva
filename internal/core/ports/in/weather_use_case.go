package in

import (
	"context"
	"weather-api/internal/core/domain"
)

type WeatherUseCase interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
}
