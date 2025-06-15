package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"weather-api/internal/core/domain"
	"weather-api/internal/mocks"
)

func TestWeatherService_GetWeather(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		city        string
		setupMocks  func(p *mocks.MockWeatherProvider)
		expected    domain.Weather
		expectedErr error
	}{
		{
			name: "returns weather successfully",
			city: "Kyiv",
			setupMocks: func(p *mocks.MockWeatherProvider) {
				p.On("GetWeather", ctx, "Kyiv").
					Return(domain.Weather{
						Temperature: 20.5,
						Humidity:    60,
						Description: "Sunny",
					}, nil).
					Once()
			},
			expected: domain.Weather{
				Temperature: 20.5,
				Humidity:    60,
				Description: "Sunny",
			},
			expectedErr: nil,
		},
		{
			name: "provider returns error",
			city: "Atlantis",
			setupMocks: func(p *mocks.MockWeatherProvider) {
				p.On("GetWeather", ctx, "Atlantis").
					Return(domain.Weather{}, errors.New("API error")).Once()
			},
			expected:    domain.Weather{},
			expectedErr: errors.New("API error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerMock := &mocks.MockWeatherProvider{}
			tt.setupMocks(providerMock)

			ws := NewWeatherService(providerMock)

			result, err := ws.GetWeather(ctx, tt.city)

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)

			providerMock.AssertExpectations(t)
		})
	}
}
