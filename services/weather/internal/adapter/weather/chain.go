package weather

import (
	"context"
	"errors"
	"weather-service/internal/core/domain"
)

type ProviderHandler struct {
	provider Provider
	next     *ProviderHandler
}

func (h *ProviderHandler) SetNext(handler *ProviderHandler) *ProviderHandler {
	h.next = handler
	return handler
}

func (h *ProviderHandler) HandleGetWeather(ctx context.Context, city string) (domain.Weather, error) {
	weather, err := h.provider.GetWeather(ctx, city)
	if err == nil {
		return weather, nil
	}

	if h.next != nil {
		return h.next.HandleGetWeather(ctx, city)
	}

	return domain.Weather{}, errors.New("all weather providers unavailable")
}

func (h *ProviderHandler) HandleCheckCityExists(ctx context.Context, city string) error {
	err := h.provider.CheckCityExists(ctx, city)
	if err == nil {
		return nil
	}

	if h.next != nil {
		return h.next.HandleCheckCityExists(ctx, city)
	}

	return errors.New("city not found in any provider")
}

type ChainWeatherProvider struct {
	start *ProviderHandler
}

func (c *ChainWeatherProvider) Name() string {
	return "ChainWeatherProvider"
}

func NewChainWeatherProvider(providers ...Provider) *ChainWeatherProvider {
	if len(providers) == 0 {
		return &ChainWeatherProvider{}
	}

	startHandler := &ProviderHandler{provider: providers[0]}
	currentHandler := startHandler

	for i := 1; i < len(providers); i++ {
		nextHandler := &ProviderHandler{provider: providers[i]}
		currentHandler.SetNext(nextHandler)
		currentHandler = nextHandler
	}

	return &ChainWeatherProvider{start: startHandler}
}

func (c *ChainWeatherProvider) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	if c.start == nil {
		return domain.Weather{}, errors.New("no weather providers in chain")
	}
	return c.start.HandleGetWeather(ctx, city)
}

func (c *ChainWeatherProvider) CheckCityExists(ctx context.Context, city string) error {
	if c.start == nil {
		return errors.New("no weather providers in chain")
	}
	return c.start.HandleCheckCityExists(ctx, city)
}
