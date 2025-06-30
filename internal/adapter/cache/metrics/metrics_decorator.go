package metrics

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	redisadapter "weather-api/internal/adapter/cache/redis"

	"weather-api/internal/core/domain"
)

type CacheWithMetrics struct {
	cache   redisadapter.WeatherCache
	metrics *CacheMetrics
}

func NewCacheWithMetrics(cache redisadapter.WeatherCache, metrics *CacheMetrics) redisadapter.WeatherCache {
	return &CacheWithMetrics{cache: cache, metrics: metrics}
}

func (c *CacheWithMetrics) Get(ctx context.Context, city string) (*domain.Weather, error) {
	if c.metrics == nil {
		return c.cache.Get(ctx, city)
	}

	start := time.Now()
	defer func() {
		c.metrics.OperationDuration.Observe(time.Since(start).Seconds())
	}()

	weather, err := c.cache.Get(ctx, city)
	if errors.Is(err, redis.Nil) {
		c.metrics.Misses.Inc()
		return nil, nil
	}
	if err != nil {
		c.metrics.Errors.Inc()
		return nil, err
	}

	c.metrics.Hits.WithLabelValues(strings.ToLower(city)).Inc()
	return weather, nil
}

func (c *CacheWithMetrics) Set(ctx context.Context, city string, value domain.Weather) error {
	if c.metrics == nil {
		return c.cache.Set(ctx, city, value)
	}

	if city == "" {
		c.metrics.Skipped.Inc()
		return nil
	}

	start := time.Now()
	defer func() {
		c.metrics.OperationDuration.Observe(time.Since(start).Seconds())
	}()

	err := c.cache.Set(ctx, city, value)
	if err != nil {
		c.metrics.Errors.Inc()
		return err
	}

	return nil
}

func (c *CacheWithMetrics) Close() error {
	return c.cache.Close()
}
