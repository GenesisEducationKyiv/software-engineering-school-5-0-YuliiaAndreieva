package service

import (
	"context"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports/out"
)

type WeatherService struct {
	provider out.WeatherProvider
}

func NewWeatherService(provider out.WeatherProvider) *WeatherService {
	return &WeatherService{
		provider: provider,
	}
}

func (s *WeatherService) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	return s.provider.GetWeather(ctx, city)
}
