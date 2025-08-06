package out

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics interface {
	WeatherRequestsTotal() prometheus.Counter
	WeatherRequestDuration() prometheus.Histogram
	WeatherRequestErrors() prometheus.Counter
	CacheHits() prometheus.Counter
	CacheMisses() prometheus.Counter
	CacheErrors() prometheus.Counter
	ActiveConnections() prometheus.Gauge
}

type WeatherMetrics struct {
	weatherRequestsTotal   prometheus.Counter
	weatherRequestDuration prometheus.Histogram
	weatherRequestErrors   prometheus.Counter
	cacheHits              prometheus.Counter
	cacheMisses            prometheus.Counter
	cacheErrors            prometheus.Counter
	activeConnections      prometheus.Gauge
}

func NewWeatherMetrics(reg prometheus.Registerer) *WeatherMetrics {
	return &WeatherMetrics{
		weatherRequestsTotal: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "weather_requests_total",
			Help: "Total number of weather requests",
		}),
		weatherRequestDuration: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "weather_request_duration_seconds",
			Help:    "Duration of weather requests",
			Buckets: prometheus.DefBuckets,
		}),
		weatherRequestErrors: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "weather_request_errors_total",
			Help: "Total number of weather request errors",
		}),
		cacheHits: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		}),
		cacheMisses: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		}),
		cacheErrors: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "cache_errors_total",
			Help: "Total number of cache errors",
		}),
		activeConnections: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		}),
	}
}

func (m *WeatherMetrics) WeatherRequestsTotal() prometheus.Counter {
	return m.weatherRequestsTotal
}

func (m *WeatherMetrics) WeatherRequestDuration() prometheus.Histogram {
	return m.weatherRequestDuration
}

func (m *WeatherMetrics) WeatherRequestErrors() prometheus.Counter {
	return m.weatherRequestErrors
}

func (m *WeatherMetrics) CacheHits() prometheus.Counter {
	return m.cacheHits
}

func (m *WeatherMetrics) CacheMisses() prometheus.Counter {
	return m.cacheMisses
}

func (m *WeatherMetrics) CacheErrors() prometheus.Counter {
	return m.cacheErrors
}

func (m *WeatherMetrics) ActiveConnections() prometheus.Gauge {
	return m.activeConnections
}
