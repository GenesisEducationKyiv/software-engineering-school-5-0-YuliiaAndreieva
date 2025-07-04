package weather

import (
	"context"
	"encoding/json"
	"errors"
	"weather-api/internal/adapter/cache/core"
	"weather-api/internal/adapter/cache/core/metrics"
	"weather-api/internal/core/domain"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, city string) (*domain.Weather, error)
	Set(ctx context.Context, city string, weather domain.Weather) error
	Close() error
}

type CacheImpl struct {
	cache metrics.Cache
}

func NewCache(cache metrics.Cache) *CacheImpl {
	return &CacheImpl{cache: cache}
}

func (w *CacheImpl) Get(ctx context.Context, city string) (*domain.Weather, error) {
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

func (w *CacheImpl) Set(ctx context.Context, city string, weather domain.Weather) error {
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

func (w *CacheImpl) Close() error {
	return w.cache.Close()
}
