package strategies

import (
	"bytes"
	"context"
	"email/internal/adapter/email/templates"
	"email/internal/core/domain"
	"email/internal/core/ports/out"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
)

type WeatherUpdateEmailStrategy struct {
	logger                 out.Logger
	subscriptionServiceURL string
}

func NewWeatherUpdateEmailStrategy(logger out.Logger, subscriptionServiceURL string) out.EmailTemplateStrategy {
	return &WeatherUpdateEmailStrategy{
		logger:                 logger,
		subscriptionServiceURL: subscriptionServiceURL,
	}
}

func (s *WeatherUpdateEmailStrategy) BuildTemplate(ctx context.Context, data interface{}) (string, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected map[string]interface{}, got %T", data)
	}

	jsonData, err := json.Marshal(dataMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	var weatherData domain.WeatherUpdateEmailData
	if err := json.Unmarshal(jsonData, &weatherData); err != nil {
		return "", fmt.Errorf("failed to unmarshal weather data: %w", err)
	}

	unsubscribeLink := ""
	if weatherData.UnsubscribeToken != "" {
		unsubscribeLink = templates.BuildUnsubscribeLink(s.subscriptionServiceURL, weatherData.UnsubscribeToken)
	}

	tmpl, err := template.New("weather").Parse(templates.WeatherUpdateEmailTemplate)
	if err != nil {
		return "", err
	}

	templateData := templates.WeatherUpdateEmailData{
		City:            weatherData.City,
		Temperature:     strconv.FormatFloat(weatherData.Temperature, 'f', 1, 64),
		Description:     weatherData.Description,
		Humidity:        strconv.FormatFloat(weatherData.Humidity, 'f', 1, 64),
		WindSpeed:       strconv.FormatFloat(weatherData.WindSpeed, 'f', 1, 64),
		UnsubscribeLink: template.HTML(unsubscribeLink),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *WeatherUpdateEmailStrategy) GetType() string {
	return string(domain.EmailTypeWeatherUpdate)
}
