package openweathermap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
	"weather-api/internal/util/jsonutil"
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

func (c *Client) Name() string {
	return "OpenWeatherMap"
}

func (c *Client) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	endpoint := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=metric", c.baseURL, url.QueryEscape(city), c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		msg := fmt.Sprintf("unable to create HTTP request for city %s: %v", city, err)
		log.Print(msg)
		return domain.Weather{}, errors.New(msg)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("unable to make HTTP request for city %s: %v", city, err)
		log.Print(msg)
		return domain.Weather{}, errors.New(msg)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	var logBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &logBuffer)

	weatherResp, err := jsonutil.Decode[Response](teeReader)
	if err != nil {
		msg := fmt.Sprintf("unable to decode openweathermap response: %v", err)
		log.Print(msg)
		return domain.Weather{}, weather.NewProviderError(c.Name(), 500, msg)
	}

	c.logger.Log(c.Name(), logBuffer.Bytes())

	if code, ok := weatherResp.Cod.(float64); ok && code != 200 {
		log.Printf("OpenWeatherMap error (code: %.0f): %s", code, weatherResp.Message)
		return domain.Weather{}, c.mapError(int(code), weatherResp.Message)
	}

	return domain.Weather{
		Temperature: weatherResp.Main.Temp,
		Humidity:    weatherResp.Main.Humidity,
		Description: weatherResp.Weather[0].Description,
	}, nil
}

func (c *Client) CheckCityExists(ctx context.Context, city string) error {
	endpoint := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=metric", c.baseURL, url.QueryEscape(city), c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		msg := fmt.Sprintf("creating request: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("sending request: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return weather.NewProviderError(c.Name(), 500, "unable to read response body")
	}

	c.logger.Log(c.Name(), bodyBytes)

	if resp.StatusCode == http.StatusNotFound {
		return weather.NewProviderError(c.Name(), 404, "City not found")
	}

	return nil
}
