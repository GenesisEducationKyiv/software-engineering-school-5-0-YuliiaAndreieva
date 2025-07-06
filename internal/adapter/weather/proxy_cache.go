package weather

import (
	"context"
	"log"
	weathercache "weather-api/internal/adapter/cache/weather"
	"weather-api/internal/core/domain"
)

type CachedWeatherProvider struct {
	cache    weathercache.Cache
	upstream Provider
}

func NewCachedWeatherProvider(cache weathercache.Cache, upstream Provider) *CachedWeatherProvider {
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
