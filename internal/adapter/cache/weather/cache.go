package weather

import (
	"context"
	"encoding/json"
	"errors"
	"weather-api/internal/adapter/cache/core"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	cache ports.Cache
}

func NewCache(cache ports.Cache) *Cache {
	return &Cache{cache: cache}
}

func (w *Cache) Get(ctx context.Context, city string) (*domain.Weather, error) {
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

func (w *Cache) Set(ctx context.Context, city string, weather domain.Weather) error {
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

func (w *Cache) Close() error {
	return w.cache.Close()
}
