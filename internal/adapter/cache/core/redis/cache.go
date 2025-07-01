package redis

import (
	"context"
	"time"
	"weather-api/internal/adapter/cache/core"

	"github.com/redis/go-redis/v9"
)

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

func NewCache(opts CacheOptions) *Cache {
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

func (r *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, core.NewError(core.InvalidKey, key, nil)
	}
	result := r.client.Get(ctx, key)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return result.Bytes()
}

func (r *Cache) Set(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return core.NewError(core.InvalidKey, key, nil)
	}
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *Cache) Close() error {
	return r.client.Close()
}
