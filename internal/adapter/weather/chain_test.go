//go:build unit
// +build unit

package weather

import (
	"context"
	"errors"
	"testing"
	"weather-api/internal/core/domain"
)

type MockFailingProvider struct{}

func (m *MockFailingProvider) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	return domain.Weather{}, errors.New("first provider error")
}

func (m *MockFailingProvider) CheckCityExists(ctx context.Context, city string) error {
	return errors.New("first provider error")
}

func (m *MockFailingProvider) Name() string {
	return "MockFailingProvider"
}

type MockSuccessfulProvider struct{}

func (m *MockSuccessfulProvider) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	return domain.Weather{
		Temperature: 25.0,
		Humidity:    70,
		Description: "Sunny",
	}, nil
}

func (m *MockSuccessfulProvider) CheckCityExists(ctx context.Context, city string) error {
	return nil
}

func (m *MockSuccessfulProvider) Name() string {
	return "MockSuccessfulProvider"
}

func TestChainWeatherProvider_Fallback(t *testing.T) {
	failingProvider := &MockFailingProvider{}
	successfulProvider := &MockSuccessfulProvider{}

	chain := NewChainWeatherProvider(failingProvider, successfulProvider)

	weather, err := chain.GetWeather(context.Background(), "Kyiv")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if weather.Temperature != 25.0 {
		t.Errorf("Expected temperature 25.0, got: %f", weather.Temperature)
	}

	if weather.Humidity != 70 {
		t.Errorf("Expected humidity 70, got: %d", weather.Humidity)
	}

	if weather.Description != "Sunny" {
		t.Errorf("Expected description 'Sunny', got: %s", weather.Description)
	}

	err = chain.CheckCityExists(context.Background(), "Kyiv")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestChainWeatherProvider_AllProvidersFail(t *testing.T) {
	failingProvider1 := &MockFailingProvider{}
	failingProvider2 := &MockFailingProvider{}

	chain := NewChainWeatherProvider(failingProvider1, failingProvider2)

	_, err := chain.GetWeather(context.Background(), "Kyiv")
	if err == nil {
		t.Fatal("Expected error when all providers fail")
	}

	if err.Error() != "all weather providers unavailable" {
		t.Errorf("Expected 'all weather providers unavailable', got: %s", err.Error())
	}

	err = chain.CheckCityExists(context.Background(), "Kyiv")
	if err == nil {
		t.Fatal("Expected error when all providers fail")
	}

	if err != domain.ErrCityNotFound {
		t.Errorf("Expected domain.ErrCityNotFound, got: %v", err)
	}
}
