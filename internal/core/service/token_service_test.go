package service

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	"weather-api/internal/core/port"
)

func TestTokenService_GenerateToken(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() port.TokenService
		verify func(t *testing.T, token string, err error)
	}{
		{
			name: "success",
			setup: func() port.TokenService {
				return NewTokenService()
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
