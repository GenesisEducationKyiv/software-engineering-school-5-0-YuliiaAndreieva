package in

import (
	"context"
	"weather-api/internal/core/ports/out"
)

type SubscribeUseCase interface {
	Subscribe(ctx context.Context, opts out.SubscribeOptions) (string, error)
}
