package unit

import (
	"testing"

	"email-service/internal/adapter/dto"
	"email-service/internal/adapter/http"

	"github.com/stretchr/testify/assert"
)

func TestEmailValidator_ValidateEmailFormat(t *testing.T) {
	validator := http.NewEmailValidator()

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "Valid email",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "Another valid email",
			email:    "user@test.org",
			expected: true,
		},
		{
			name:     "Invalid email - no @",
			email:    "testexample.com",
			expected: false,
		},
		{
			name:     "Invalid email - no dot",
			email:    "test@example",
			expected: false,
		},
		{
			name:     "Empty email",
			email:    "",
			expected: false,
		},
		{
			name:     "Invalid email format",
			email:    "invalid-email",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEmailFormat(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmailValidator_ValidateRequiredFields(t *testing.T) {
	validator := http.NewEmailValidator()

	tests := []struct {
		name     string
		fields   map[string]string
		expected []string
	}{
		{
			name: "All fields present",
			fields: map[string]string{
				"To":   "test@example.com",
				"City": "Kyiv",
			},
			expected: []string{},
		},
		{
			name: "Empty field",
			fields: map[string]string{
				"To":   "test@example.com",
				"City": "",
			},
			expected: []string{"City is required"},
		},
		{
			name: "Multiple empty fields",
			fields: map[string]string{
				"To":   "",
				"City": "",
				"Name": "John",
			},
			expected: []string{"To is required", "City is required"},
		},
		{
			name: "Whitespace only fields",
			fields: map[string]string{
				"To":   "test@example.com",
				"City": "   ",
			},
			expected: []string{"City is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateRequiredFields(tt.fields)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestEmailValidator_ValidateConfirmationEmailRequest(t *testing.T) {
	validator := http.NewEmailValidator()

	tests := []struct {
		name     string
		request  dto.ConfirmationEmailRequest
		expected []string
	}{
		{
			name: "Valid request",
			request: dto.ConfirmationEmailRequest{
				To:               "test@example.com",
				Subject:          "Confirm Subscription",
				City:             "Kyiv",
				ConfirmationLink: "http://localhost/confirm/token123",
			},
			expected: []string{},
		},
		{
			name: "Invalid email format",
			request: dto.ConfirmationEmailRequest{
				To:               "invalid-email",
				Subject:          "Confirm Subscription",
				City:             "Kyiv",
				ConfirmationLink: "http://localhost/confirm/token123",
			},
			expected: []string{"Invalid email format"},
		},
		{
			name: "Missing required fields",
			request: dto.ConfirmationEmailRequest{
				To:               "",
				Subject:          "",
				City:             "",
				ConfirmationLink: "",
			},
			expected: []string{"Invalid email format", "To is required", "Subject is required", "City is required", "ConfirmationLink is required"},
		},
		{
			name: "Partial missing fields",
			request: dto.ConfirmationEmailRequest{
				To:               "test@example.com",
				Subject:          "",
				City:             "Kyiv",
				ConfirmationLink: "http://localhost/confirm/token123",
			},
			expected: []string{"Subject is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfirmationEmailRequest(tt.request)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestEmailValidator_ValidateWeatherUpdateEmailRequest(t *testing.T) {
	validator := http.NewEmailValidator()

	tests := []struct {
		name     string
		request  dto.WeatherUpdateEmailRequest
		expected []string
	}{
		{
			name: "Valid request",
			request: dto.WeatherUpdateEmailRequest{
				To:          "test@example.com",
				Subject:     "Weather Update",
				Name:        "User",
				City:        "Kyiv",
				Temperature: 15,
				Description: "Partly cloudy",
				Humidity:    65,
				WindSpeed:   12,
			},
			expected: []string{},
		},
		{
			name: "Invalid email format",
			request: dto.WeatherUpdateEmailRequest{
				To:          "invalid-email",
				Subject:     "Weather Update",
				Name:        "User",
				City:        "Kyiv",
				Temperature: 15,
				Description: "Partly cloudy",
				Humidity:    65,
				WindSpeed:   12,
			},
			expected: []string{"Invalid email format"},
		},
		{
			name: "Missing required fields",
			request: dto.WeatherUpdateEmailRequest{
				To:          "",
				Subject:     "",
				Name:        "",
				City:        "",
				Temperature: 15,
				Description: "",
				Humidity:    65,
				WindSpeed:   12,
			},
			expected: []string{"Invalid email format", "To is required", "Subject is required", "Name is required", "City is required", "Description is required"},
		},
		{
			name: "Partial missing fields",
			request: dto.WeatherUpdateEmailRequest{
				To:          "test@example.com",
				Subject:     "Weather Update",
				Name:        "",
				City:        "Kyiv",
				Temperature: 15,
				Description: "Partly cloudy",
				Humidity:    65,
				WindSpeed:   12,
			},
			expected: []string{"Name is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateWeatherUpdateEmailRequest(tt.request)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
