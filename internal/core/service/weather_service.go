package service

import (
	"context"
	"log"
	weathercache "weather-api/internal/adapter/cache/weather"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
)

type WeatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
}

type weatherService struct {
	weatherSvc weather.Provider
	cache      weathercache.Cache
}

func NewWeatherService(weatherSvc weather.Provider, cache weathercache.Cache) WeatherService {
	return &weatherService{
		weatherSvc: weatherSvc,
		cache:      cache,
	}
}

func (s *weatherService) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	if cached, err := s.cache.Get(ctx, city); err == nil && cached != nil {
		return *cached, nil
	}

	data, err := s.weatherSvc.GetWeather(ctx, city)
	if err != nil {
		return domain.Weather{}, err
	}

	if err := s.cache.Set(ctx, city, data); err != nil {
		log.Printf("cache weather for city %q: %v", city, err)
	}

	return data, nil
}
