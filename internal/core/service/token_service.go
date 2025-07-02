package service

import (
	"crypto/rand"
	"encoding/base64"
)

type TokenService interface {
	GenerateToken() (string, error)
}
type tokenService struct{}

func NewTokenService() TokenService {
	return &tokenService{}
}

func (s *tokenService) GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
