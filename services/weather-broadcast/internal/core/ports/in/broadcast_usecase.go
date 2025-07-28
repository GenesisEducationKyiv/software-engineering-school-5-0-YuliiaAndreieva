package in

import (
	"context"
	"weather-broadcast-service/internal/core/domain"
)

type BroadcastUseCase interface {
	Broadcast(ctx context.Context, frequency domain.Frequency) error
}
