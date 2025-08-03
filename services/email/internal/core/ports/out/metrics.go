package out

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics interface {
	EmailRequestsTotal() prometheus.Counter
	EmailRequestDuration() prometheus.Histogram
	EmailRequestErrors() prometheus.Counter
	EmailsSent() prometheus.Counter
	EmailsFailed() prometheus.Counter
	MessageBrokerEventsConsumed() prometheus.Counter
	MessageBrokerEventsFailed() prometheus.Counter
	SMTPConnectionsActive() prometheus.Gauge
}

type EmailMetrics struct {
	emailRequestsTotal          prometheus.Counter
	emailRequestDuration        prometheus.Histogram
	emailRequestErrors          prometheus.Counter
	emailsSent                  prometheus.Counter
	emailsFailed                prometheus.Counter
	messageBrokerEventsConsumed prometheus.Counter
	messageBrokerEventsFailed   prometheus.Counter
	smtpConnectionsActive       prometheus.Gauge
}

func NewEmailMetrics(reg prometheus.Registerer) *EmailMetrics {
	return &EmailMetrics{
		emailRequestsTotal: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "email_requests_total",
			Help: "Total number of email requests",
		}),
		emailRequestDuration: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "email_request_duration_seconds",
			Help:    "Duration of email requests",
			Buckets: prometheus.DefBuckets,
		}),
		emailRequestErrors: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "email_request_errors_total",
			Help: "Total number of email request errors",
		}),
		emailsSent: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "emails_sent_total",
			Help: "Total number of emails sent",
		}),
		emailsFailed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "emails_failed_total",
			Help: "Total number of emails failed",
		}),
		messageBrokerEventsConsumed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "message_broker_events_consumed_total",
			Help: "Total number of message broker events consumed",
		}),
		messageBrokerEventsFailed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "message_broker_events_failed_total",
			Help: "Total number of message broker events failed",
		}),
		smtpConnectionsActive: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "smtp_connections_active",
			Help: "Number of active SMTP connections",
		}),
	}
}

func (m *EmailMetrics) EmailRequestsTotal() prometheus.Counter {
	return m.emailRequestsTotal
}

func (m *EmailMetrics) EmailRequestDuration() prometheus.Histogram {
	return m.emailRequestDuration
}

func (m *EmailMetrics) EmailRequestErrors() prometheus.Counter {
	return m.emailRequestErrors
}

func (m *EmailMetrics) EmailsSent() prometheus.Counter {
	return m.emailsSent
}

func (m *EmailMetrics) EmailsFailed() prometheus.Counter {
	return m.emailsFailed
}

func (m *EmailMetrics) MessageBrokerEventsConsumed() prometheus.Counter {
	return m.messageBrokerEventsConsumed
}

func (m *EmailMetrics) MessageBrokerEventsFailed() prometheus.Counter {
	return m.messageBrokerEventsFailed
}

func (m *EmailMetrics) SMTPConnectionsActive() prometheus.Gauge {
	return m.smtpConnectionsActive
}
