package weather

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"weather-api/internal/adapter/cache/core"
	"weather-api/internal/adapter/cache/core/metrics"
	"weather-api/internal/core/domain"
)

type Cache interface {
	Get(ctx context.Context, city string) (*domain.Weather, error)
	Set(ctx context.Context, city string, weather domain.Weather) error
	Close() error
}

type CacheIml struct {
	cache metrics.Cache
}

func NewCache(cache metrics.Cache) *CacheIml {
	return &CacheIml{cache: cache}
}

func (w *CacheIml) Get(ctx context.Context, city string) (*domain.Weather, error) {
	if city == "" {
		return nil, core.NewError(core.InvalidKey, city, nil)
	}
	data, err := w.cache.Get(ctx, city)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, core.NewError(core.RedisError, city, err)
	}
	if data == nil {
		return nil, nil
	}
	var weather domain.Weather
	if err := json.Unmarshal(data, &weather); err != nil {
		return nil, core.NewError(core.UnmarshalError, city, err)
	}
	return &weather, nil
}

func (w *CacheIml) Set(ctx context.Context, city string, weather domain.Weather) error {
	if city == "" {
		return core.NewError(core.InvalidKey, city, nil)
	}
	data, err := json.Marshal(weather)
	if err != nil {
		return core.NewError(core.MarshalError, city, err)
	}
	if err := w.cache.Set(ctx, city, data); err != nil {
		return core.NewError(core.RedisError, city, err)
	}
	return nil
}

func (w *CacheIml) Close() error {
	return w.cache.Close()
}
