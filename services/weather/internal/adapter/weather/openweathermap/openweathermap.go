package openweathermap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"weather-service/internal/adapter/weather"
	"weather-service/internal/core/domain"
)

const weatherEndpoint = "/weather"

type Client struct {
	apiKey     string
	baseURL    string
	httpClient weather.HTTPDoer
	logger     weather.ProviderLogger
}

type ClientOptions struct {
	APIKey     string
	BaseURL    string
	HTTPClient weather.HTTPDoer
	Logger     weather.ProviderLogger
}

func NewClient(opts ClientOptions) *Client {
	return &Client{
		apiKey:     opts.APIKey,
		baseURL:    opts.BaseURL,
		httpClient: opts.HTTPClient,
		logger:     opts.Logger,
	}
}

func (c *Client) Name() string {
	return "OpenWeatherMap"
}

func (c *Client) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	req, err := c.createRequest(ctx, city)
	if err != nil {
		return domain.Weather{}, err
	}

	resp, err := weather.ExecuteRequest(c.httpClient, req)
	if err != nil {
		return domain.Weather{}, err
	}
	defer weather.CloseResponse(resp)

	weatherResp, responseBytes, err := weather.DecodeResponse[Response](resp)
	if err != nil {
		return domain.Weather{}, err
	}

	c.logger.Log(c.Name(), responseBytes)

	if code, ok := weatherResp.Cod.(float64); ok && code != 200 {
		return domain.Weather{}, c.mapError(int(code), weatherResp.Message)
	}

	return convertToDomain(weatherResp), nil
}

func (c *Client) CheckCityExists(ctx context.Context, city string) error {
	req, err := c.createRequest(ctx, city)
	if err != nil {
		return err
	}

	resp, err := weather.ExecuteRequest(c.httpClient, req)
	if err != nil {
		return err
	}
	defer weather.CloseResponse(resp)

	if resp.StatusCode == http.StatusNotFound {
		return c.mapError(404, "city not found")
	}

	return nil
}

func (c *Client) createRequest(ctx context.Context, city string) (*http.Request, error) {
	requestURL := fmt.Sprintf("%s%s?q=%s&appid=%s&units=metric", c.baseURL, weatherEndpoint, url.QueryEscape(city), c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		msg := fmt.Sprintf("creating HTTP request for city %s: %v", city, err)
		return nil, errors.New(msg)
	}
	return req, nil
}
