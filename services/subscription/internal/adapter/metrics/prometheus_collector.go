package metrics

import (
	"subscription/internal/core/ports/out"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusCollector struct {
	httpRequestsTotal        *prometheus.CounterVec
	httpDuration             *prometheus.HistogramVec
	subscriptionCreated      prometheus.Counter
	subscriptionConfirmed    prometheus.Counter
	subscriptionUnsubscribed prometheus.Counter
	subscriptionErrors       prometheus.Counter
	rabbitMQPublished        prometheus.Counter
	rabbitMQPublishErrors    prometheus.Counter
	databaseOperations       *prometheus.HistogramVec
	databaseErrors           prometheus.Counter
	grpcRequestsTotal        *prometheus.CounterVec
	grpcErrorsTotal          *prometheus.CounterVec
	grpcDuration             *prometheus.HistogramVec
}

func NewPrometheusCollector() out.MetricsCollector {
	return NewPrometheusCollectorWithOptions(
		WithHTTPRequestsTotal(),
		WithHTTPDuration(),
		WithSubscriptionCreated(),
		WithSubscriptionConfirmed(),
		WithSubscriptionUnsubscribed(),
		WithSubscriptionErrors(),
		WithRabbitMQPublished(),
		WithRabbitMQPublishErrors(),
		WithDatabaseOperations(),
		WithDatabaseErrors(),
		WithGRPCRequestsTotal(),
		WithGRPCErrorsTotal(),
		WithGRPCDuration(),
	)
}

func (p *PrometheusCollector) IncrementHTTPRequests(method, path string, statusCode int) {
	if p.httpRequestsTotal != nil {
		p.httpRequestsTotal.WithLabelValues(method, path, string(rune(statusCode))).Inc()
	}
}

func (p *PrometheusCollector) RecordHTTPDuration(method, path string, duration float64) {
	if p.httpDuration != nil {
		p.httpDuration.WithLabelValues(method, path).Observe(duration)
	}
}

func (p *PrometheusCollector) IncrementSubscriptionCreated() {
	p.subscriptionCreated.Inc()
}

func (p *PrometheusCollector) IncrementSubscriptionConfirmed() {
	p.subscriptionConfirmed.Inc()
}

func (p *PrometheusCollector) IncrementSubscriptionUnsubscribed() {
	p.subscriptionUnsubscribed.Inc()
}

func (p *PrometheusCollector) IncrementSubscriptionErrors() {
	p.subscriptionErrors.Inc()
}

func (p *PrometheusCollector) IncrementRabbitMQPublished() {
	p.rabbitMQPublished.Inc()
}

func (p *PrometheusCollector) IncrementRabbitMQPublishErrors() {
	p.rabbitMQPublishErrors.Inc()
}

func (p *PrometheusCollector) RecordDatabaseOperation(operation string, duration float64) {
	if p.databaseOperations != nil {
		p.databaseOperations.WithLabelValues(operation).Observe(duration)
	}
}

func (p *PrometheusCollector) IncrementDatabaseErrors() {
	p.databaseErrors.Inc()
}

func (p *PrometheusCollector) IncrementGRPCRequests(service, method string) {
	if p.grpcRequestsTotal != nil {
		p.grpcRequestsTotal.WithLabelValues(service, method).Inc()
	}
}

func (p *PrometheusCollector) IncrementGRPCErrors(service, method string) {
	if p.grpcErrorsTotal != nil {
		p.grpcErrorsTotal.WithLabelValues(service, method).Inc()
	}
}

func (p *PrometheusCollector) RecordGRPCDuration(service, method string, duration float64) {
	if p.grpcDuration != nil {
		p.grpcDuration.WithLabelValues(service, method).Observe(duration)
	}
}
