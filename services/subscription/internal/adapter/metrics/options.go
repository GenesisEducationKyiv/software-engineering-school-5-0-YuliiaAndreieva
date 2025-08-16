package metrics

import (
	"subscription/internal/core/ports/out"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsOption func(*PrometheusCollector)

func WithHTTPRequestsTotal() MetricsOption {
	return func(p *PrometheusCollector) {
		p.httpRequestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		)
	}
}

func WithHTTPDuration() MetricsOption {
	return func(p *PrometheusCollector) {
		p.httpDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		)
	}
}

func WithSubscriptionCreated() MetricsOption {
	return func(p *PrometheusCollector) {
		p.subscriptionCreated = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "subscription_created_total",
				Help: "Total number of subscriptions created",
			},
		)
	}
}

func WithSubscriptionConfirmed() MetricsOption {
	return func(p *PrometheusCollector) {
		p.subscriptionConfirmed = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "subscription_confirmed_total",
				Help: "Total number of subscriptions confirmed",
			},
		)
	}
}

func WithSubscriptionUnsubscribed() MetricsOption {
	return func(p *PrometheusCollector) {
		p.subscriptionUnsubscribed = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "subscription_unsubscribed_total",
				Help: "Total number of subscriptions unsubscribed",
			},
		)
	}
}

func WithSubscriptionErrors() MetricsOption {
	return func(p *PrometheusCollector) {
		p.subscriptionErrors = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "subscription_errors_total",
				Help: "Total number of subscription errors",
			},
		)
	}
}

func WithRabbitMQPublished() MetricsOption {
	return func(p *PrometheusCollector) {
		p.rabbitMQPublished = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "rabbitmq_published_total",
				Help: "Total number of RabbitMQ messages published",
			},
		)
	}
}

func WithRabbitMQPublishErrors() MetricsOption {
	return func(p *PrometheusCollector) {
		p.rabbitMQPublishErrors = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "rabbitmq_publish_errors_total",
				Help: "Total number of RabbitMQ publish errors",
			},
		)
	}
}

func WithDatabaseOperations() MetricsOption {
	return func(p *PrometheusCollector) {
		p.databaseOperations = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_operation_duration_seconds",
				Help:    "Database operation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		)
	}
}

func WithDatabaseErrors() MetricsOption {
	return func(p *PrometheusCollector) {
		p.databaseErrors = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "database_errors_total",
				Help: "Total number of database errors",
			},
		)
	}
}

func WithGRPCRequestsTotal() MetricsOption {
	return func(p *PrometheusCollector) {
		p.grpcRequestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"service", "method"},
		)
	}
}

func WithGRPCErrorsTotal() MetricsOption {
	return func(p *PrometheusCollector) {
		p.grpcErrorsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_errors_total",
				Help: "Total number of gRPC errors",
			},
			[]string{"service", "method"},
		)
	}
}

func WithGRPCDuration() MetricsOption {
	return func(p *PrometheusCollector) {
		p.grpcDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		)
	}
}

func NewPrometheusCollectorWithOptions(options ...MetricsOption) out.MetricsCollector {
	collector := &PrometheusCollector{}

	for _, option := range options {
		option(collector)
	}

	return collector
}
