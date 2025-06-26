package repository

import (
	"context"
	"weather-api/internal/core/domain"
)

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub domain.Subscription) error
	GetSubscriptionByToken(ctx context.Context, token string) (domain.Subscription, error)
	UpdateSubscription(ctx context.Context, sub domain.Subscription) error
	DeleteSubscription(ctx context.Context, token string) error
	GetSubscriptionsByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error)
	IsTokenExists(ctx context.Context, token string) (bool, error)
	IsSubscriptionExists(ctx context.Context, opts IsSubscriptionExistsOptions) (bool, error)
}

type CityRepository interface {
	Create(ctx context.Context, city domain.City) (domain.City, error)
	GetByName(ctx context.Context, name string) (domain.City, error)
}

type IsSubscriptionExistsOptions struct {
	Email     string
	CityID    int64
	Frequency domain.Frequency
}

type SubscribeOptions struct {
	Email     string
	City      string
	Frequency domain.Frequency
}
