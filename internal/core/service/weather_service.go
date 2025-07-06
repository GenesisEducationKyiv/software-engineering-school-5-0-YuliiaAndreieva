package service

import (
	"context"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
)

type WeatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
}

type weatherService struct {
	provider weather.Provider
}

func NewWeatherService(provider weather.Provider) WeatherService {
	return &weatherService{
		provider: provider,
	}
}

func (s *weatherService) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	return s.provider.GetWeather(ctx, city)
}
