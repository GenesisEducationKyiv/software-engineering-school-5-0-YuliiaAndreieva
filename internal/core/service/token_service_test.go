//go:build unit
// +build unit

package service

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"weather-api/internal/core/domain"
	"weather-api/internal/mocks"
)

func TestTokenService_GenerateToken(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() TokenService
		verify func(t *testing.T, token string, err error)
	}{
		{
			name: "success",
			setup: func() TokenService {
				repo := &mocks.MockSubscriptionRepository{}
				return NewTokenService(repo)
			},
			verify: func(t *testing.T, token string, err error) {
				assert.NoError(t, err, "GenerateToken should not return an error")

				assert.NotEmpty(t, token, "Token should not be empty")

				decoded, err := base64.URLEncoding.DecodeString(token)
				assert.NoError(t, err, "Token should be valid base64 URL-encoded")
				assert.Len(t, decoded, 32, "Decoded token should be 32 bytes")

				assert.Len(t, token, 44, "Encoded token should be 44 characters")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()

			token, err := svc.GenerateToken()

			tt.verify(t, token, err)
		})
	}
}

func TestTokenService_CheckTokenExists(t *testing.T) {
	ctx := context.Background()
	const token = "test-token"

	tests := []struct {
		name       string
		token      string
		setupMocks func(repo *mocks.MockSubscriptionRepository)
		expectErr  error
	}{
		{
			name:  "token exists",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository) {
				repo.On("IsTokenExists", ctx, token).Return(true, nil)
			},
			expectErr: nil,
		},
		{
			name:  "token not found",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository) {
				repo.On("IsTokenExists", ctx, token).Return(false, nil)
			},
			expectErr: domain.ErrTokenNotFound,
		},
		{
			name:  "empty token",
			token: "",
			setupMocks: func(repo *mocks.MockSubscriptionRepository) {
				// No mock needed for empty token
			},
			expectErr: domain.ErrInvalidToken,
		},
		{
			name:  "repository error",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository) {
				repo.On("IsTokenExists", ctx, token).Return(false, errors.New("db error"))
			},
			expectErr: errors.New("unable to check token existence: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			tt.setupMocks(repo)

			svc := NewTokenService(repo)

			err := svc.CheckTokenExists(ctx, tt.token)
			assert.Equal(t, tt.expectErr, err)
			repo.AssertExpectations(t)
		})
	}
}
