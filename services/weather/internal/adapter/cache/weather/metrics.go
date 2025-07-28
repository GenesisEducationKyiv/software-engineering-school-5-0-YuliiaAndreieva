package weather

import (
	"github.com/prometheus/client_golang/prometheus"
	"weather-service/internal/adapter/cache/core/metrics"
)

func NewCacheMetrics(reg prometheus.Registerer) *metrics.CacheMetrics {
	return metrics.NewCacheMetrics(reg, "weather")
}
