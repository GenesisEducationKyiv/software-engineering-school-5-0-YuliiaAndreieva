package request

import (
	"strings"
	"weather-api/internal/handler/http/errors"
)

type TokenRequest struct {
	Token string
}

func NewTokenRequest(token string) *TokenRequest {
	return &TokenRequest{Token: token}
}

func (r *TokenRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return errors.ErrTokenRequired
	}

	return nil
}
