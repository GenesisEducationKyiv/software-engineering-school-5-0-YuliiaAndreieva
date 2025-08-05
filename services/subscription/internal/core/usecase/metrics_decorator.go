package usecase

import (
	"context"
	"time"

	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"
)

type SubscribeWithMetrics struct {
	base    in.SubscribeUseCase
	metrics out.MetricsCollector
}

func NewSubscribeWithMetrics(base in.SubscribeUseCase, metrics out.MetricsCollector) in.SubscribeUseCase {
	return &SubscribeWithMetrics{
		base:    base,
		metrics: metrics,
	}
}

func (d *SubscribeWithMetrics) Subscribe(ctx context.Context, req domain.SubscriptionRequest) (*domain.SubscriptionResponse, error) {
	start := time.Now()

	result, err := d.base.Subscribe(ctx, req)

	duration := time.Since(start).Seconds()
	d.metrics.RecordGRPCDuration("subscription", "subscribe", duration)

	if err != nil {
		d.metrics.IncrementGRPCErrors("subscription", "subscribe")
		d.metrics.IncrementSubscriptionErrors()
	} else {
		d.metrics.IncrementGRPCRequests("subscription", "subscribe")
		if result.Success {
			d.metrics.IncrementSubscriptionCreated()
		}
	}

	return result, err
}

type ConfirmSubscriptionWithMetrics struct {
	base    in.ConfirmSubscriptionUseCase
	metrics out.MetricsCollector
}

func NewConfirmSubscriptionWithMetrics(base in.ConfirmSubscriptionUseCase, metrics out.MetricsCollector) in.ConfirmSubscriptionUseCase {
	return &ConfirmSubscriptionWithMetrics{
		base:    base,
		metrics: metrics,
	}
}

func (d *ConfirmSubscriptionWithMetrics) ConfirmSubscription(ctx context.Context, token string) (*domain.ConfirmResponse, error) {
	start := time.Now()

	result, err := d.base.ConfirmSubscription(ctx, token)

	duration := time.Since(start).Seconds()
	d.metrics.RecordGRPCDuration("subscription", "confirm", duration)

	if err != nil {
		d.metrics.IncrementGRPCErrors("subscription", "confirm")
		d.metrics.IncrementSubscriptionErrors()
	} else {
		d.metrics.IncrementGRPCRequests("subscription", "confirm")
		if result.Success {
			d.metrics.IncrementSubscriptionConfirmed()
		}
	}

	return result, err
}

type UnsubscribeWithMetrics struct {
	base    in.UnsubscribeUseCase
	metrics out.MetricsCollector
}

func NewUnsubscribeWithMetrics(base in.UnsubscribeUseCase, metrics out.MetricsCollector) in.UnsubscribeUseCase {
	return &UnsubscribeWithMetrics{
		base:    base,
		metrics: metrics,
	}
}

func (d *UnsubscribeWithMetrics) Unsubscribe(ctx context.Context, req domain.UnsubscribeRequest) (*domain.UnsubscribeResponse, error) {
	start := time.Now()

	result, err := d.base.Unsubscribe(ctx, req)

	duration := time.Since(start).Seconds()
	d.metrics.RecordGRPCDuration("subscription", "unsubscribe", duration)

	if err != nil {
		d.metrics.IncrementGRPCErrors("subscription", "unsubscribe")
		d.metrics.IncrementSubscriptionErrors()
	} else {
		d.metrics.IncrementGRPCRequests("subscription", "unsubscribe")
		if result.Success {
			d.metrics.IncrementSubscriptionUnsubscribed()
		}
	}

	return result, err
}
