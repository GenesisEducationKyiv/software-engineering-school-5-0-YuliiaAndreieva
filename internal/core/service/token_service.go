package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"
)

type TokenServiceImpl struct {
	repo ports.SubscriptionRepository
}

func NewTokenService(repo ports.SubscriptionRepository) *TokenServiceImpl {
	return &TokenServiceImpl{
		repo: repo,
	}
}

func (s *TokenServiceImpl) GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *TokenServiceImpl) CheckTokenExists(ctx context.Context, token string) error {
	if token == "" {
		log.Printf("Invalid token provided: empty token")
		return domain.ErrInvalidToken
	}

	exists, err := s.repo.IsTokenExists(ctx, token)
	if err != nil {
		msg := fmt.Sprintf("unable to check token existence: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}
	if !exists {
		log.Printf("Token not found: %s", token)
		return domain.ErrTokenNotFound
	}

	return nil
}
