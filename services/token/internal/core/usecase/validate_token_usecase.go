package usecase

import (
	"context"
	sharedlogger "shared/logger"
	"token/internal/core/domain"
	"token/internal/core/ports/in"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email,omitempty"`
}

type ValidateTokenUseCase struct {
	logger sharedlogger.Logger
	secret []byte
}

func NewValidateTokenUseCase(logger sharedlogger.Logger, secret string) in.ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		logger: logger,
		secret: []byte(secret),
	}
}

func (uc *ValidateTokenUseCase) ValidateToken(ctx context.Context, req domain.ValidateTokenRequest) (*domain.ValidateTokenResponse, error) {
	uc.logger.Infof("Validating JWT token")

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return uc.secret, nil
	}

	var claims CustomClaims
	token, err := jwt.ParseWithClaims(req.Token, &claims, keyFunc)

	if err != nil {
		uc.logger.Warnf("Token validation failed: %v", err)
		return &domain.ValidateTokenResponse{
			Valid:   false,
			Message: "Token is invalid",
			Error:   err.Error(),
		}, nil
	}

	if token.Valid {
		uc.logger.Infof("Token validated successfully for subject: %s", claims.Subject)
		return &domain.ValidateTokenResponse{
			Valid:   true,
			Message: "Token is valid",
		}, nil
	}

	uc.logger.Warnf("Token is invalid")
	return &domain.ValidateTokenResponse{
		Valid:   false,
		Message: "Token is invalid",
	}, nil
}
