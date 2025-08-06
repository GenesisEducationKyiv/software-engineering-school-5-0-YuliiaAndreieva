package mocks

import (
	"context"
	"time"

	"weather/internal/core/domain"
)

type MockWeatherProvider struct {
	getWeatherFunc      func(ctx context.Context, city string) (domain.Weather, error)
	checkCityExistsFunc func(ctx context.Context, city string) error
	nameFunc            func() string
}

func NewMockWeatherProvider() *MockWeatherProvider {
	return &MockWeatherProvider{}
}

func (m *MockWeatherProvider) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	if m.getWeatherFunc != nil {
		return m.getWeatherFunc(ctx, city)
	}
	return domain.Weather{
		City:        city,
		Temperature: 20.0,
		Humidity:    65,
		Description: "Partly cloudy",
		WindSpeed:   10.0,
		Timestamp:   time.Now(),
	}, nil
}

func (m *MockWeatherProvider) CheckCityExists(ctx context.Context, city string) error {
	if m.checkCityExistsFunc != nil {
		return m.checkCityExistsFunc(ctx, city)
	}
	return nil
}

func (m *MockWeatherProvider) Name() string {
	if m.nameFunc != nil {
		return m.nameFunc()
	}
	return "MockWeatherProvider"
}

func (m *MockWeatherProvider) WithGetWeather(fn func(ctx context.Context, city string) (domain.Weather, error)) *MockWeatherProvider {
	m.getWeatherFunc = fn
	return m
}

func (m *MockWeatherProvider) WithCheckCityExists(fn func(ctx context.Context, city string) error) *MockWeatherProvider {
	m.checkCityExistsFunc = fn
	return m
}

func (m *MockWeatherProvider) WithName(fn func() string) *MockWeatherProvider {
	m.nameFunc = fn
	return m
}

type MockChainWeatherProvider struct {
	getWeatherFunc      func(ctx context.Context, city string) (domain.Weather, error)
	checkCityExistsFunc func(ctx context.Context, city string) error
	nameFunc            func() string
}

func NewMockChainWeatherProvider() *MockChainWeatherProvider {
	return &MockChainWeatherProvider{}
}

func (m *MockChainWeatherProvider) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	if m.getWeatherFunc != nil {
		return m.getWeatherFunc(ctx, city)
	}
	return domain.Weather{
		City:        city,
		Temperature: 20.0,
		Humidity:    65,
		Description: "Partly cloudy",
		WindSpeed:   10.0,
		Timestamp:   time.Now(),
	}, nil
}

func (m *MockChainWeatherProvider) CheckCityExists(ctx context.Context, city string) error {
	if m.checkCityExistsFunc != nil {
		return m.checkCityExistsFunc(ctx, city)
	}
	return nil
}

func (m *MockChainWeatherProvider) Name() string {
	if m.nameFunc != nil {
		return m.nameFunc()
	}
	return "MockChainWeatherProvider"
}

func (m *MockChainWeatherProvider) WithGetWeather(fn func(ctx context.Context, city string) (domain.Weather, error)) *MockChainWeatherProvider {
	m.getWeatherFunc = fn
	return m
}

func (m *MockChainWeatherProvider) WithCheckCityExists(fn func(ctx context.Context, city string) error) *MockChainWeatherProvider {
	m.checkCityExistsFunc = fn
	return m
}

func (m *MockChainWeatherProvider) WithName(fn func() string) *MockChainWeatherProvider {
	m.nameFunc = fn
	return m
}
