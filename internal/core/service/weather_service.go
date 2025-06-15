package service

import (
	"weather-api/internal/core/domain"
)

type WeatherService interface {
	GetWeather(city string) (domain.Weather, error)
}

type weatherService struct {
	weatherSvc WeatherService
}

func NewWeatherService(weatherSvc WeatherService) *weatherService {
	return &weatherService{weatherSvc: weatherSvc}
}

func (s *weatherService) GetWeather(city string) (domain.Weather, error) {
	weather, err := s.weatherSvc.GetWeather(city)
	if err != nil {
		return domain.Weather{}, err
	}
	return weather, nil
}
