package usecase

import (
	"context"
	"time"

	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"
)

type SubscribeMetricsDecorator struct {
	base    in.SubscribeUseCase
	metrics out.MetricsCollector
}

func NewSubscribeMetricsDecorator(base in.SubscribeUseCase, metrics out.MetricsCollector) in.SubscribeUseCase {
	return &SubscribeMetricsDecorator{
		base:    base,
		metrics: metrics,
	}
}

func (d *SubscribeMetricsDecorator) Subscribe(ctx context.Context, req domain.SubscriptionRequest) (*domain.SubscriptionResponse, error) {
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

type ConfirmSubscriptionMetricsDecorator struct {
	base    in.ConfirmSubscriptionUseCase
	metrics out.MetricsCollector
}

func NewConfirmSubscriptionMetricsDecorator(base in.ConfirmSubscriptionUseCase, metrics out.MetricsCollector) in.ConfirmSubscriptionUseCase {
	return &ConfirmSubscriptionMetricsDecorator{
		base:    base,
		metrics: metrics,
	}
}

func (d *ConfirmSubscriptionMetricsDecorator) ConfirmSubscription(ctx context.Context, token string) (*domain.ConfirmResponse, error) {
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

type UnsubscribeMetricsDecorator struct {
	base    in.UnsubscribeUseCase
	metrics out.MetricsCollector
}

func NewUnsubscribeMetricsDecorator(base in.UnsubscribeUseCase, metrics out.MetricsCollector) in.UnsubscribeUseCase {
	return &UnsubscribeMetricsDecorator{
		base:    base,
		metrics: metrics,
	}
}

func (d *UnsubscribeMetricsDecorator) Unsubscribe(ctx context.Context, req domain.UnsubscribeRequest) (*domain.UnsubscribeResponse, error) {
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
