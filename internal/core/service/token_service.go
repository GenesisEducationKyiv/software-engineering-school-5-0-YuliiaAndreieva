package service

import (
	"crypto/rand"
	"encoding/base64"
	"weather-api/internal/core/port"
)

type TokenService struct{}

func NewTokenService() port.TokenService {
	return &TokenService{}
}

func (s *TokenService) GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
