package in

import (
	"context"
	"weather-service/internal/core/domain"
)

type GetWeatherUseCase interface {
	GetWeather(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error)
}
