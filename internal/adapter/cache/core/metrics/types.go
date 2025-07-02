package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CacheMetrics struct {
	Hits              *prometheus.CounterVec
	Misses            prometheus.Counter
	Errors            prometheus.Counter
	Skipped           prometheus.Counter
	OperationDuration prometheus.Histogram
	reg               prometheus.Registerer
	namespace         string
}

type CacheMetricsOption func(*CacheMetrics)

func WithHits() CacheMetricsOption {
	return func(m *CacheMetrics) {
		m.Hits = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: m.namespace,
			Name:      "cache_hits",
			Help:      "Cache hits per key",
		}, []string{"key"})
		m.reg.MustRegister(m.Hits)
	}
}

func WithMisses() CacheMetricsOption {
	return func(m *CacheMetrics) {
		m.Misses = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.namespace,
			Name:      "cache_misses_total",
			Help:      "Total number of cache misses",
		})
		m.reg.MustRegister(m.Misses)
	}
}

func WithErrors() CacheMetricsOption {
	return func(m *CacheMetrics) {
		m.Errors = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.namespace,
			Name:      "cache_errors_total",
			Help:      "Total number of cache errors",
		})
		m.reg.MustRegister(m.Errors)
	}
}

func WithSkipped() CacheMetricsOption {
	return func(m *CacheMetrics) {
		m.Skipped = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.namespace,
			Name:      "cache_skipped_total",
			Help:      "Number of skipped or ignored cache sets",
		})
		m.reg.MustRegister(m.Skipped)
	}
}

func WithOperationDuration() CacheMetricsOption {
	return func(m *CacheMetrics) {
		m.OperationDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: m.namespace,
			Name:      "cache_operation_duration_seconds",
			Help:      "Duration of cache operations",
			Buckets:   prometheus.DefBuckets,
		})
		m.reg.MustRegister(m.OperationDuration)
	}
}

func NewCacheMetrics(reg prometheus.Registerer, namespace string, opts ...CacheMetricsOption) *CacheMetrics {
	m := &CacheMetrics{
		reg:       reg,
		namespace: namespace,
	}

	if len(opts) == 0 {
		opts = []CacheMetricsOption{
			WithHits(),
			WithMisses(),
			WithErrors(),
			WithSkipped(),
			WithOperationDuration(),
		}
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}
