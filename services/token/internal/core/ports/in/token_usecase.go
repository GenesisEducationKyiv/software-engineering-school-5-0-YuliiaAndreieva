package in

import (
	"context"
	"token/internal/core/domain"
)

//go:generate mockery --name GenerateTokenUseCase
type GenerateTokenUseCase interface {
	GenerateToken(ctx context.Context, req domain.GenerateTokenRequest) (*domain.GenerateTokenResponse, error)
}

//go:generate mockery --name ValidateTokenUseCase
type ValidateTokenUseCase interface {
	ValidateToken(ctx context.Context, req domain.ValidateTokenRequest) (*domain.ValidateTokenResponse, error)
}
