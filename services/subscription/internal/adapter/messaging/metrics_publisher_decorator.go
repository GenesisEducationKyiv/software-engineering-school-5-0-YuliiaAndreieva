package messaging

import (
	"context"
	"time"

	"subscription/internal/core/domain"
	"subscription/internal/core/ports/out"
)

type RabbitMQMetricsPublisherDecorator struct {
	base    out.EventPublisher
	metrics out.MetricsCollector
}

func NewRabbitMQMetricsPublisherDecorator(base out.EventPublisher, metrics out.MetricsCollector) out.EventPublisher {
	return &RabbitMQMetricsPublisherDecorator{
		base:    base,
		metrics: metrics,
	}
}

func (d *RabbitMQMetricsPublisherDecorator) PublishSubscriptionCreated(ctx context.Context, subscription domain.Subscription) error {
	start := time.Now()

	err := d.base.PublishSubscriptionCreated(ctx, subscription)

	duration := time.Since(start).Seconds()
	d.metrics.RecordDatabaseOperation("rabbitmq_publish", duration)

	if err != nil {
		d.metrics.IncrementRabbitMQPublishErrors()
		d.metrics.IncrementSubscriptionErrors()
	} else {
		d.metrics.IncrementRabbitMQPublished()
	}

	return err
}
