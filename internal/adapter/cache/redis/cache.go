package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"weather-api/internal/adapter/cache"
	"weather-api/internal/core/domain"

	"github.com/redis/go-redis/v9"
)

type WeatherCache interface {
	Get(ctx context.Context, city string) (*domain.Weather, error)
	Set(ctx context.Context, city string, value domain.Weather) error
	Close() error
}

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

type CacheOptions struct {
	Address      string
	TTL          time.Duration
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
}

func New(opts CacheOptions) WeatherCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:         opts.Address,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
	})
	return &Cache{
		client: rdb,
		ttl:    opts.TTL,
	}
}

func (c *Cache) Get(ctx context.Context, city string) (*domain.Weather, error) {
	key := c.buildKey(city)
	if key == "" {
		return nil, cache.NewError(cache.InvalidKey, key, nil)
	}
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, cache.NewError(cache.RedisError, key, err)
	}

	var weather domain.Weather
	if err := json.Unmarshal([]byte(val), &weather); err != nil {
		return nil, cache.NewError(cache.UnmarshalError, key, err)
	}
	return &weather, nil
}

func (c *Cache) Set(ctx context.Context, city string, value domain.Weather) error {
	key := c.buildKey(city)
	if key == "" {
		return cache.NewError(cache.InvalidKey, key, nil)
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return cache.NewError(cache.MarshalError, key, err)
	}

	if err := c.client.Set(ctx, key, bytes, c.ttl).Err(); err != nil {
		return cache.NewError(cache.RedisError, key, err)
	}
	return nil
}

func (c *Cache) buildKey(city string) string {
	if city == "" {
		return ""
	}
	return fmt.Sprintf("weather:city%s", strings.ToLower(city))
}

func (c *Cache) Close() error {
	return c.client.Close()
}
