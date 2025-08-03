package out

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics interface {
	SubscriptionRequestsTotal() prometheus.Counter
	SubscriptionRequestDuration() prometheus.Histogram
	SubscriptionRequestErrors() prometheus.Counter
	SubscriptionsCreated() prometheus.Counter
	SubscriptionsConfirmed() prometheus.Counter
	SubscriptionsUnsubscribed() prometheus.Counter
	MessageBrokerEventsPublished() prometheus.Counter
	MessageBrokerEventsFailed() prometheus.Counter
	DatabaseOperationsTotal() prometheus.Counter
	DatabaseOperationDuration() prometheus.Histogram
}

type SubscriptionMetrics struct {
	subscriptionRequestsTotal    prometheus.Counter
	subscriptionRequestDuration  prometheus.Histogram
	subscriptionRequestErrors    prometheus.Counter
	subscriptionsCreated         prometheus.Counter
	subscriptionsConfirmed       prometheus.Counter
	subscriptionsUnsubscribed    prometheus.Counter
	messageBrokerEventsPublished prometheus.Counter
	messageBrokerEventsFailed    prometheus.Counter
	databaseOperationsTotal      prometheus.Counter
	databaseOperationDuration    prometheus.Histogram
}

func NewSubscriptionMetrics(reg prometheus.Registerer) *SubscriptionMetrics {
	return &SubscriptionMetrics{
		subscriptionRequestsTotal: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "subscription_requests_total",
			Help: "Total number of subscription requests",
		}),
		subscriptionRequestDuration: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "subscription_request_duration_seconds",
			Help:    "Duration of subscription requests",
			Buckets: prometheus.DefBuckets,
		}),
		subscriptionRequestErrors: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "subscription_request_errors_total",
			Help: "Total number of subscription request errors",
		}),
		subscriptionsCreated: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "subscriptions_created_total",
			Help: "Total number of subscriptions created",
		}),
		subscriptionsConfirmed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "subscriptions_confirmed_total",
			Help: "Total number of subscriptions confirmed",
		}),
		subscriptionsUnsubscribed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "subscriptions_unsubscribed_total",
			Help: "Total number of subscriptions unsubscribed",
		}),
		messageBrokerEventsPublished: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "message_broker_events_published_total",
			Help: "Total number of message broker events published",
		}),
		messageBrokerEventsFailed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "message_broker_events_failed_total",
			Help: "Total number of message broker events failed",
		}),
		databaseOperationsTotal: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "Total number of database operations",
		}),
		databaseOperationDuration: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "database_operation_duration_seconds",
			Help:    "Duration of database operations",
			Buckets: prometheus.DefBuckets,
		}),
	}
}

func (m *SubscriptionMetrics) SubscriptionRequestsTotal() prometheus.Counter {
	return m.subscriptionRequestsTotal
}

func (m *SubscriptionMetrics) SubscriptionRequestDuration() prometheus.Histogram {
	return m.subscriptionRequestDuration
}

func (m *SubscriptionMetrics) SubscriptionRequestErrors() prometheus.Counter {
	return m.subscriptionRequestErrors
}

func (m *SubscriptionMetrics) SubscriptionsCreated() prometheus.Counter {
	return m.subscriptionsCreated
}

func (m *SubscriptionMetrics) SubscriptionsConfirmed() prometheus.Counter {
	return m.subscriptionsConfirmed
}

func (m *SubscriptionMetrics) SubscriptionsUnsubscribed() prometheus.Counter {
	return m.subscriptionsUnsubscribed
}

func (m *SubscriptionMetrics) MessageBrokerEventsPublished() prometheus.Counter {
	return m.messageBrokerEventsPublished
}

func (m *SubscriptionMetrics) MessageBrokerEventsFailed() prometheus.Counter {
	return m.messageBrokerEventsFailed
}

func (m *SubscriptionMetrics) DatabaseOperationsTotal() prometheus.Counter {
	return m.databaseOperationsTotal
}

func (m *SubscriptionMetrics) DatabaseOperationDuration() prometheus.Histogram {
	return m.databaseOperationDuration
}
