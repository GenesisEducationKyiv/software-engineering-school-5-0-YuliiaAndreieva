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
	weatherSvc weather.Provider
}

func NewWeatherService(weatherSvc weather.Provider) *weatherService {
	return &weatherService{weatherSvc: weatherSvc}
}

func (s *weatherService) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	weather, err := s.weatherSvc.GetWeather(ctx, city)
	if err != nil {
		return domain.Weather{}, err
	}
	return weather, nil
}
