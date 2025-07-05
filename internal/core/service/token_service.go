package service

import (
	"crypto/rand"
	"encoding/base64"
)

type TokenServiceImpl struct{}

func NewTokenService() *TokenServiceImpl {
	return &TokenServiceImpl{}
}

func (s *TokenServiceImpl) GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
