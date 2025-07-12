package in

import (
	"context"
)

type ConfirmSubscriptionUseCase interface {
	ConfirmSubscription(ctx context.Context, token string) error
}
