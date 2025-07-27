package token

import (
	"crypto/rand"
	"encoding/hex"
	"subscription-service/internal/core/ports/out"
)

type TokenService struct {
	logger out.Logger
}

func NewTokenService(logger out.Logger) out.TokenService {
	return &TokenService{
		logger: logger,
	}
}

func (ts *TokenService) GenerateToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)
	ts.logger.Infof("Generated new token: %s", token)
	return token
}

func (ts *TokenService) ValidateToken(token string) bool {
	if len(token) != 32 {
		ts.logger.Warnf("Invalid token length: %d", len(token))
		return false
	}
	
	_, err := hex.DecodeString(token)
	if err != nil {
		ts.logger.Warnf("Invalid token format: %v", err)
		return false
	}
	
	ts.logger.Infof("Token validated successfully: %s", token)
	return true
} 