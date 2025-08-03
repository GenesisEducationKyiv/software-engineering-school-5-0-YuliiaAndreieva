package weatherapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"weather/internal/adapter/
	"weather/internal/adapter/weather"
	"time"
)

const (
	weatherEndpoint = "/current.json"
	searchEndpoint  = "/search.json"
)

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

func apiToDomain(w response) domain.Weather {
	return domain.Weather{
		Temperature: w.TempC,
		Humidity:    w.Humidity,
		Description: w.Condition.Text,
		WindSpeed:   w.WindKph,
		Timestamp:   time.Now(),
	}
}

func (c *Client) Name() string {
	return "WeatherAPI"
}

func (c *Client) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	req, err := c.createRequest(ctx, city, weatherEndpoint)
	if err != nil {
		return domain.Weather{}, err
	}

	resp, err := weather.ExecuteRequest(c.httpClient, req)
	if err != nil {
		return domain.Weather{}, err
	}
	defer weather.CloseResponse(resp)

	env, responseBytes, err := weather.DecodeResponse[currentEnvelope](resp)
	c.logger.Log(c.Name(), responseBytes)
	if err != nil {
		return domain.Weather{}, err
	}

	if env.Error.Code != 0 {
		return domain.Weather{}, c.mapError(env.Error.Code, env.Error.Message)
	}

	return apiToDomain(env.Current), nil
}

func (c *Client) CheckCityExists(ctx context.Context, city string) error {
	req, err := c.createRequest(ctx, city, searchEndpoint)
	if err != nil {
		return err
	}

	resp, err := weather.ExecuteRequest(c.httpClient, req)
	if err != nil {
		return err
	}
	defer weather.CloseResponse(resp)

	results, responseBytes, err := weather.DecodeResponse[[]searchItem](resp)
	c.logger.Log(c.Name(), responseBytes)
	if err != nil {
		return err
	}

	if len(*results) == 0 {
		return c.mapError(404, "city not found")
	}

	return nil
}

func (c *Client) createRequest(ctx context.Context, city, endpoint string) (*http.Request, error) {
	requestURL := fmt.Sprintf("%s%s?key=%s&q=%s", c.baseURL, endpoint, c.apiKey, url.QueryEscape(city))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		msg := fmt.Sprintf("creating HTTP request for city %s: %v", city, err)
		return nil, errors.New(msg)
	}
	return req, nil
}
