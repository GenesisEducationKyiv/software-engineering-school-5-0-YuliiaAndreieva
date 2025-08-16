package in

import (
	"context"
	"weather-broadcast/internal/core/domain"
)

//go:generate mockery --name BroadcastUseCase
type BroadcastUseCase interface {
	Broadcast(ctx context.Context, frequency domain.Frequency) error
}
