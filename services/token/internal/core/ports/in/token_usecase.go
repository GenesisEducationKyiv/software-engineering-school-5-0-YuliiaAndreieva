package in

import (
	"context"
	"token/internal/core/domain"
)

type GenerateTokenUseCase interface {
	GenerateToken(ctx context.Context, req domain.GenerateTokenRequest) (*domain.GenerateTokenResponse, error)
}

type ValidateTokenUseCase interface {
	ValidateToken(ctx context.Context, req domain.ValidateTokenRequest) (*domain.ValidateTokenResponse, error)
}
