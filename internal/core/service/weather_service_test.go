package service

import (
	"errors"
	"testing"
	"weather-api/internal/mocks"

	"github.com/stretchr/testify/assert"

	"weather-api/internal/core/domain"
)

func TestWeatherService_GetWeather(t *testing.T) {
	tests := []struct {
		name          string
		city          string
		setupMocks    func(weatherSvc *mocks.MockWeatherService)
		verifyMocks   func(t *testing.T, weatherSvc *mocks.MockWeatherService)
		expected      domain.Weather
		expectedError error
	}{
		{
			name: "success",
			city: "Kyiv",
			setupMocks: func(weatherSvc *mocks.MockWeatherService) {
				weather := domain.Weather{
					Temperature: 20.5,
					Humidity:    60,
					Description: "Sunny",
				}
				weatherSvc.On("GetWeather", "Kyiv").Return(weather, nil)
			},
			verifyMocks: func(t *testing.T, weatherSvc *mocks.MockWeatherService) {
				weatherSvc.AssertExpectations(t)
			},
			expected: domain.Weather{
				Temperature: 20.5,
				Humidity:    60,
				Description: "Sunny",
			},
			expectedError: nil,
		},
		{
			name: "error from weather service",
			city: "InvalidCity",
			setupMocks: func(weatherSvc *mocks.MockWeatherService) {
				weatherSvc.On("GetWeather", "InvalidCity").Return(domain.Weather{}, errors.New("API error"))
			},
			verifyMocks: func(t *testing.T, weatherSvc *mocks.MockWeatherService) {
				weatherSvc.AssertExpectations(t)
			},
			expected:      domain.Weather{},
			expectedError: errors.New("API error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weatherSvc := &mocks.MockWeatherService{}
			service := NewWeatherService(weatherSvc)

			tt.setupMocks(weatherSvc)

			result, err := service.GetWeather(tt.city)

			assert.Equal(t, tt.expected, result, "Weather result should match expected")
			assert.Equal(t, tt.expectedError, err, "Error should match expected")
			tt.verifyMocks(t, weatherSvc)
		})
	}
}
