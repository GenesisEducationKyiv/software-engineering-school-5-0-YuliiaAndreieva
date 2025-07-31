package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"
	"token/internal/core/domain"
	"token/internal/core/ports/in"
	"token/internal/core/ports/out"

	"github.com/golang-jwt/jwt/v5"
)

type GenerateTokenUseCase struct {
	logger out.Logger
	secret []byte
}

func NewGenerateTokenUseCase(logger out.Logger, secret string) in.GenerateTokenUseCase {
	return &GenerateTokenUseCase{
		logger: logger,
		secret: []byte(secret),
	}
}

func (uc *GenerateTokenUseCase) GenerateToken(ctx context.Context, req domain.GenerateTokenRequest) (*domain.GenerateTokenResponse, error) {
	uc.logger.Infof("Generating JWT token for subject: %s", req.Subject)

	expiresIn := 24 * time.Hour
	if req.ExpiresIn != "" {
		if parsed, err := time.ParseDuration(req.ExpiresIn); err == nil {
			expiresIn = parsed
		} else {
			uc.logger.Warnf("Invalid expires_in format: %s, using default 24h", req.ExpiresIn)
		}
	}

	claims := jwt.MapClaims{
		"sub": req.Subject,
		"exp": time.Now().Add(expiresIn).Unix(),
		"iat": time.Now().Unix(),
		"jti": uc.generateJTI(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(uc.secret)
	if err != nil {
		uc.logger.Errorf("Failed to sign token: %v", err)
		return &domain.GenerateTokenResponse{
			Success: false,
			Message: "Failed to generate token",
			Error:   err.Error(),
		}, nil
	}

	uc.logger.Infof("Successfully generated JWT token for subject: %s", req.Subject)
	return &domain.GenerateTokenResponse{
		Success: true,
		Token:   tokenString,
		Message: "Token generated successfully",
	}, nil
}

func (uc *GenerateTokenUseCase) generateJTI() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		uc.logger.Errorf("Failed to generate random bytes: %v", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
