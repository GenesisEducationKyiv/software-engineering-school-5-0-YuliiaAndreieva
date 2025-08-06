package weatherapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
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
	if err != nil {
		return domain.Weather{}, err
	}

	c.logger.Log(c.Name(), responseBytes)

	if env.Error.Code != 0 {
		log.Printf("API Error detected: %v", env.Error)
		return domain.Weather{}, c.mapError(env.Error.Code, env.Error.Message)
	}

	return apiToDomain(env.Current), nil
}

func (c *Client) CheckCityExists(ctx context.Context, city string) error {
	log.Printf("Checking if city exists: %s", city)

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
	if err != nil {
		return err
	}

	c.logger.Log(c.Name(), responseBytes)

	if len(*results) == 0 {
		log.Printf("City not found in WeatherAPI database")
		return c.mapError(404, "city not found")
	}

	log.Printf("City %s exists in WeatherAPI database", city)
	return nil
}

func (c *Client) createRequest(ctx context.Context, city, endpoint string) (*http.Request, error) {
	requestURL := fmt.Sprintf("%s%s?key=%s&q=%s", c.baseURL, endpoint, c.apiKey, url.QueryEscape(city))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		msg := fmt.Sprintf("creating HTTP request for city %s: %v", city, err)
		c.logger.Log(c.Name(), []byte(msg))
		return nil, errors.New(msg)
	}
	return req, nil
}
