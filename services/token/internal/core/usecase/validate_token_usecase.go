package usecase

import (
	"context"
	"token/internal/core/domain"
	"token/internal/core/ports/in"
	"token/internal/core/ports/out"

	"github.com/golang-jwt/jwt/v5"
)

type ValidateTokenUseCase struct {
	logger out.Logger
	secret []byte
}

func NewValidateTokenUseCase(logger out.Logger, secret string) in.ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		logger: logger,
		secret: []byte(secret),
	}
}

func (uc *ValidateTokenUseCase) ValidateToken(ctx context.Context, req domain.ValidateTokenRequest) (*domain.ValidateTokenResponse, error) {
	uc.logger.Infof("Validating JWT token")

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return uc.secret, nil
	})

	if err != nil {
		uc.logger.Warnf("Token validation failed: %v", err)
		return &domain.ValidateTokenResponse{
			Success: true,
			Valid:   false,
			Message: "Token is invalid",
			Error:   err.Error(),
		}, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uc.logger.Infof("Token validated successfully for subject: %s", claims["sub"])
		return &domain.ValidateTokenResponse{
			Success: true,
			Valid:   true,
			Message: "Token is valid",
		}, nil
	}

	uc.logger.Warnf("Token is invalid")
	return &domain.ValidateTokenResponse{
		Success: true,
		Valid:   false,
		Message: "Token is invalid",
	}, nil
}
