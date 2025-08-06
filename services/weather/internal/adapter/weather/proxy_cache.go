package weather

import (
	"context"
	"log"
	"weather/internal/core/domain"
	"weather/internal/core/ports/out"
)

type Cache interface {
	Get(ctx context.Context, city string) (*domain.Weather, error)
	Set(ctx context.Context, city string, weather domain.Weather) error
	Close() error
}
type CachedWeatherProvider struct {
	cache    Cache
	upstream out.WeatherProvider
}

func NewCachedWeatherProvider(cache Cache, upstream out.WeatherProvider) *CachedWeatherProvider {
	return &CachedWeatherProvider{cache: cache, upstream: upstream}
}

func (c *CachedWeatherProvider) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	if cached, err := c.cache.Get(ctx, city); err == nil && cached != nil {
		return *cached, nil
	}
	data, err := c.upstream.GetWeather(ctx, city)
	if err != nil {
		return domain.Weather{}, err
	}
	if err := c.cache.Set(ctx, city, data); err != nil {
		log.Printf("cache weather for city %q: %v", city, err)
	}
	return data, nil
}

func (c *CachedWeatherProvider) CheckCityExists(ctx context.Context, city string) error {
	return c.upstream.CheckCityExists(ctx, city)
}

func (c *CachedWeatherProvider) Name() string {
	return c.upstream.Name()
}
