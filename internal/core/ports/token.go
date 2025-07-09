package ports

import "context"

type TokenService interface {
	GenerateToken() (string, error)
	CheckTokenExists(ctx context.Context, token string) error
}
