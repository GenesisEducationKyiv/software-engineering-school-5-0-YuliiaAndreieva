package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CacheMetrics struct {
	Hits              *prometheus.CounterVec
	Misses            prometheus.Counter
	Errors            prometheus.Counter
	Skipped           prometheus.Counter
	KeyCount          prometheus.Gauge
	OperationDuration prometheus.Histogram
}

func NewCacheMetrics(reg prometheus.Registerer) *CacheMetrics {
	m := &CacheMetrics{
		Hits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "weather_cache_hits",
			Help: "Cache hits per city",
		}, []string{"city"}),

		Misses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_misses_total",
			Help: "Total number of cache misses",
		}),

		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_errors_total",
			Help: "Total number of cache errors",
		}),

		Skipped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "weather_cache_skipped_total",
			Help: "Number of skipped or ignored cache sets",
		}),

		KeyCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "weather_cache_key_count",
			Help: "Current number of keys in Redis",
		}),

		OperationDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "weather_cache_operation_duration_seconds",
			Help:    "Duration of cache operations",
			Buckets: prometheus.DefBuckets,
		}),
	}

	reg.MustRegister(m.Hits, m.Misses, m.Errors, m.Skipped, m.KeyCount, m.OperationDuration)

	return m
}
