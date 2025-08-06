package weather

import (
	"weather/internal/adapter/cache/core/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func NewCacheMetrics(reg prometheus.Registerer) *metrics.CacheMetrics {
	return metrics.NewCacheMetrics(reg, "weather")
}
