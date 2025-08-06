//go:build integration
// +build integration

package integration

import (
	"context"
	"errors"
	"github.com/alicebob/miniredis/v2"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	corecache "weather-api/internal/adapter/cache/core"
	"weather-api/internal/adapter/cache/core/redis"
	weathercache "weather-api/internal/adapter/cache/weather"
)

func defaultRedisOptions() redis.CacheOptions {
	return redis.CacheOptions{
		Address:      "localhost:6379",
		TTL:          2 * time.Second,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     2,
		MinIdleConns: 1,
	}
}

func newMiniRedisCache(t *testing.T) (*redis.Cache, func()) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	opts := defaultRedisOptions()
	opts.Address = s.Addr()
	cache := redis.NewCache(opts)
	return cache, func() { cache.Close(); s.Close() }
}

func newWeatherCache(t *testing.T) (*weathercache.Cache, func()) {
	raw, cleanup := newMiniRedisCache(t)
	cache := weathercache.NewCache(raw)
	return cache, cleanup
}

func newWeatherCacheWithMiniRedis(t *testing.T) (*weathercache.Cache, *redis.Cache, *miniredis.Miniredis, func()) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	opts := defaultRedisOptions()
	opts.Address = s.Addr()
	raw := redis.NewCache(opts)
	cache := weathercache.NewCache(raw)
	cleanup := func() { raw.Close(); s.Close() }
	return cache, raw, s, cleanup
}

func TestRedisCache_Flow(t *testing.T) {
	cache, cleanup := newMiniRedisCache(t)
	defer cleanup()
	ctx := context.Background()
	key := "city"
	value := []byte("weather-data")

	err := cache.Set(ctx, key, value)
	require.NoError(t, err)

	got, err := cache.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, got)

	value2 := []byte("weather-data-2")
	err = cache.Set(ctx, key, value2)
	require.NoError(t, err)
	got, err = cache.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value2, got)

	s := cache
	s.Close()
}

func TestRedisCache_Miss(t *testing.T) {
	cache, cleanup := newMiniRedisCache(t)
	defer cleanup()
	ctx := context.Background()
	_, err := cache.Get(ctx, "no_such_key")
	require.ErrorIs(t, err, redisv9.Nil)
}

func TestRedisCache_InvalidKey(t *testing.T) {
	cache, cleanup := newMiniRedisCache(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cache.Get(ctx, "")
	require.Error(t, err)
	err = cache.Set(ctx, "", []byte("data"))
	require.Error(t, err)
}

func TestRedisCache_ConnectionError(t *testing.T) {
	opts := defaultRedisOptions()
	opts.Address = "localhost:9999"
	cache := redis.NewCache(opts)
	ctx := context.Background()
	_, err := cache.Get(ctx, "any")
	require.Error(t, err)
}

func TestWeatherCache_Miss(t *testing.T) {
	cache, cleanup := newWeatherCache(t)
	defer cleanup()
	ctx := context.Background()
	got, err := cache.Get(ctx, "no_such_key")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestWeatherCache_InvalidKey(t *testing.T) {
	cache, cleanup := newWeatherCache(t)
	defer cleanup()
	ctx := context.Background()
	_, err := cache.Get(ctx, "")
	var cacheErr *corecache.Error
	require.Error(t, err)
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, corecache.InvalidKey, cacheErr.Code)
}

func TestWeatherCache_UnmarshalError(t *testing.T) {
	cache, raw, _, cleanup := newWeatherCacheWithMiniRedis(t)
	defer cleanup()
	ctx := context.Background()
	city := "BadJsonCity"
	err := raw.Set(ctx, city, []byte("not-a-json"))
	require.NoError(t, err)
	_, err = cache.Get(ctx, city)
	var cacheErr *corecache.Error
	require.Error(t, err)
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, corecache.UnmarshalError, cacheErr.Code)
}

func TestWeatherCache_RedisError(t *testing.T) {
	opts := defaultRedisOptions()
	opts.Address = "localhost:9999"
	raw := redis.NewCache(opts)
	cache := weathercache.NewCache(raw)
	ctx := context.Background()
	_, err := cache.Get(ctx, "any")
	var cacheErr *corecache.Error
	require.Error(t, err)
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, corecache.RedisError, cacheErr.Code)
}
