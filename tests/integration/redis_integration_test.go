//go:build integration
// +build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"
	cachepkg "weather-api/internal/adapter/cache"
	"weather-api/internal/adapter/cache/redis"
	"weather-api/internal/core/domain"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func newMiniRedisCache(t *testing.T) (redis.WeatherCache, *miniredis.Miniredis, func()) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	cache := redis.New(redis.CacheOptions{
		Address:      s.Addr(),
		TTL:          2 * time.Second,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     2,
		MinIdleConns: 1,
	})
	return cache, s, func() { cache.Close(); s.Close() }
}

func TestRedisWeatherCache_Basic(t *testing.T) {
	cache, _, cleanup := newMiniRedisCache(t)
	defer cleanup()

	ctx := context.Background()
	city := "TestCity"
	weather := domain.Weather{
		Temperature: 25.5,
		Humidity:    50,
		Description: "Sunny",
	}

	err := cache.Set(ctx, city, weather)
	require.NoError(t, err)

	got, err := cache.Get(ctx, city)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, weather, *got)

	weather2 := domain.Weather{
		Temperature: 10.0,
		Humidity:    80,
		Description: "Rainy",
	}
	err = cache.Set(ctx, city, weather2)
	require.NoError(t, err)
	got, err = cache.Get(ctx, city)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, weather2, *got)

	got, err = cache.Get(ctx, "NoSuchCity")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestRedisWeatherCache_InvalidKey(t *testing.T) {
	cache, _, cleanup := newMiniRedisCache(t)
	defer cleanup()

	ctx := context.Background()
	_, err := cache.Get(ctx, "")
	require.Error(t, err)
	var cacheErr *cachepkg.Error
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, cachepkg.InvalidKey, cacheErr.Code)

	err = cache.Set(ctx, "", domain.Weather{})
	require.Error(t, err)
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, cachepkg.InvalidKey, cacheErr.Code)
}

func TestRedisWeatherCache_UnmarshalError(t *testing.T) {
	cache, s, cleanup := newMiniRedisCache(t)
	defer cleanup()

	ctx := context.Background()
	city := "BadJsonCity"
	key := "weather:citybadjsoncity"

	s.Set(key, "not-a-json")
	t.Log("miniredis keys:", s.Keys())

	_, err := cache.Get(ctx, city)
	require.Error(t, err)
	var cacheErr *cachepkg.Error
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, cachepkg.UnmarshalError, cacheErr.Code)
}

func TestRedisWeatherCache_RedisError(t *testing.T) {
	cache := redis.New(redis.CacheOptions{
		Address:      "localhost:9999",
		TTL:          2 * time.Second,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolSize:     2,
		MinIdleConns: 1,
	})
	ctx := context.Background()
	_, err := cache.Get(ctx, "Kyiv")
	require.Error(t, err)
	var cacheErr *cachepkg.Error
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, cachepkg.RedisError, cacheErr.Code)
	err = cache.Set(ctx, "Kyiv", domain.Weather{})
	require.Error(t, err)
	require.True(t, errors.As(err, &cacheErr))
	require.Equal(t, cachepkg.RedisError, cacheErr.Code)
}
