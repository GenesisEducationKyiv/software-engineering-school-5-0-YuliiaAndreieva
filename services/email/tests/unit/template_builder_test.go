package unit

import (
	"context"
	"testing"

	"email/internal/adapter/email"
	"email/internal/adapter/logger"
)

func TestTemplateBuilder_BuildConfirmationEmail(t *testing.T) {
	logger := logger.NewLogrusLogger()
	builder := email.NewTemplateBuilder(logger)

	tests := []struct {
		name             string
		email            string
		city             string
		confirmationLink string
		expectedContains []string
	}{
		{
			name:             "Valid confirmation email",
			email:            "test@example.com",
			city:             "Kyiv",
			confirmationLink: "http://localhost:8082/confirm/token123",
			expectedContains: []string{
				"Welcome!",
				"Thank you for subscribing to weather updates for Kyiv",
				"http://localhost:8082/confirm/token123",
				"If you didn't create this subscription",
			},
		},
		{
			name:             "Another city",
			email:            "user@test.com",
			city:             "Lviv",
			confirmationLink: "http://localhost:8082/confirm/token456",
			expectedContains: []string{
				"Thank you for subscribing to weather updates for Lviv",
				"http://localhost:8082/confirm/token456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := builder.BuildConfirmationEmail(context.Background(), tt.email, tt.city, tt.confirmationLink)

			if err != nil {
				t.Errorf("BuildConfirmationEmail() error = %v", err)
				return
			}

			if template == "" {
				t.Error("BuildConfirmationEmail() returned empty template")
				return
			}

			for _, expected := range tt.expectedContains {
				if !contains(template, expected) {
					t.Errorf("Template does not contain expected text: %s", expected)
				}
			}
		})
	}
}

func TestTemplateBuilder_BuildWeatherUpdateEmail(t *testing.T) {
	logger := logger.NewLogrusLogger()
	builder := email.NewTemplateBuilder(logger)

	tests := []struct {
		name             string
		email            string
		city             string
		description      string
		humidity         int
		windSpeed        int
		temperature      int
		expectedContains []string
	}{
		{
			name:        "Valid weather update email",
			email:       "test@example.com",
			city:        "Kyiv",
			description: "Partly cloudy",
			humidity:    65,
			windSpeed:   12,
			temperature: 15,
			expectedContains: []string{
				"Weather Update for Kyiv",
				"Hello,",
				"Temperature: 15°C",
				"Description: Partly cloudy",
				"Humidity: 65%",
				"Wind Speed: 12 km/h",
			},
		},
		{
			name:        "Another weather condition",
			email:       "user@test.com",
			city:        "Lviv",
			description: "Sunny",
			humidity:    45,
			windSpeed:   8,
			temperature: 22,
			expectedContains: []string{
				"Weather Update for Lviv",
				"Temperature: 22°C",
				"Description: Sunny",
				"Humidity: 45%",
				"Wind Speed: 8 km/h",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := builder.BuildWeatherUpdateEmail(
				context.Background(),
				tt.email,
				tt.city,
				tt.description,
				tt.humidity,
				tt.windSpeed,
				tt.temperature,
				"test-unsubscribe-token",
			)

			if err != nil {
				t.Errorf("BuildWeatherUpdateEmail() error = %v", err)
				return
			}

			if template == "" {
				t.Error("BuildWeatherUpdateEmail() returned empty template")
				return
			}

			for _, expected := range tt.expectedContains {
				if !contains(template, expected) {
					t.Errorf("Template does not contain expected text: %s", expected)
				}
			}
		})
	}
}

func TestTemplateBuilder_EdgeCases(t *testing.T) {
	logger := logger.NewLogrusLogger()
	builder := email.NewTemplateBuilder(logger)

	t.Run("Empty city", func(t *testing.T) {
		template, err := builder.BuildConfirmationEmail(context.Background(), "test@example.com", "", "http://localhost/confirm/token")

		if err != nil {
			t.Errorf("BuildConfirmationEmail() with empty city should not return error: %v", err)
		}

		if template == "" {
			t.Error("BuildConfirmationEmail() should return template even with empty city")
		}
	})

	t.Run("Special characters in city", func(t *testing.T) {
		template, err := builder.BuildConfirmationEmail(context.Background(), "test@example.com", "New York", "http://localhost/confirm/token")

		if err != nil {
			t.Errorf("BuildConfirmationEmail() with special characters should not return error: %v", err)
		}

		if !contains(template, "New York") {
			t.Error("Template should contain the city name with spaces")
		}
	})

	t.Run("Zero values in weather", func(t *testing.T) {
		template, err := builder.BuildWeatherUpdateEmail(
			context.Background(),
			"test@example.com",
			"Kyiv",
			"Clear",
			0,
			0,
			0,
			"test-unsubscribe-token",
		)

		if err != nil {
			t.Errorf("BuildWeatherUpdateEmail() with zero values should not return error: %v", err)
		}

		if !contains(template, "Temperature: 0°C") {
			t.Error("Template should handle zero temperature correctly")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
