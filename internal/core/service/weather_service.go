package service

import (
	"context"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"
)

type WeatherService struct {
	provider ports.WeatherProvider
}

func NewWeatherService(provider ports.WeatherProvider) *WeatherService {
	return &WeatherService{
		provider: provider,
	}
}

func (s *WeatherService) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	return s.provider.GetWeather(ctx, city)
}
