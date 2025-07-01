package weather

import (
	"github.com/prometheus/client_golang/prometheus"
	"weather-api/internal/adapter/cache/core/metrics"
)

func NewCacheMetrics(reg prometheus.Registerer) *metrics.CacheMetrics {
	return metrics.NewCacheMetrics(reg, "weather")
}
