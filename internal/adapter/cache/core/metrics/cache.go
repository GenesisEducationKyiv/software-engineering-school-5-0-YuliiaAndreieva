package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Close() error
}

type CacheWithMetrics struct {
	cache   Cache
	metrics *CacheMetrics
}

func NewCacheWithMetrics(cache Cache, metrics *CacheMetrics) Cache {
	return &CacheWithMetrics{
		cache:   cache,
		metrics: metrics,
	}
}

func (c *CacheWithMetrics) Get(ctx context.Context, key string) ([]byte, error) {
	if c.metrics == nil {
		return c.cache.Get(ctx, key)
	}

	start := time.Now()
	defer func() {
		c.metrics.OperationDuration.Observe(time.Since(start).Seconds())
	}()

	data, err := c.cache.Get(ctx, key)
	if errors.Is(err, redis.Nil) {
		c.metrics.Misses.Inc()
		return nil, nil
	}
	if err != nil {
		c.metrics.Errors.Inc()
		return nil, err
	}

	c.metrics.Hits.WithLabelValues(key).Inc()
	return data, nil
}

func (c *CacheWithMetrics) Set(ctx context.Context, key string, value []byte) error {
	if c.metrics == nil {
		return c.cache.Set(ctx, key, value)
	}

	if key == "" {
		c.metrics.Skipped.Inc()
		return nil
	}

	start := time.Now()
	defer func() {
		c.metrics.OperationDuration.Observe(time.Since(start).Seconds())
	}()

	err := c.cache.Set(ctx, key, value)
	if err != nil {
		c.metrics.Errors.Inc()
		return err
	}

	return nil
}

func (c *CacheWithMetrics) Close() error {
	return c.cache.Close()
}
