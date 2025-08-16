package out

import (
	"context"
)

type TokenService interface {
	GenerateToken(ctx context.Context, email string, expiresIn string) (string, error)
	ValidateToken(ctx context.Context, token string) (bool, error)
}
