package email

import (
	"bytes"
	"context"
	"html/template"
	"strconv"

	"email/internal/core/ports/out"
)

type TemplateBuilder struct {
	logger                 out.Logger
	baseURL                string
	subscriptionServiceURL string
}

func NewTemplateBuilder(logger out.Logger, baseURL string, subscriptionServiceURL string) out.EmailTemplateBuilder {
	return &TemplateBuilder{
		logger:                 logger,
		baseURL:                baseURL,
		subscriptionServiceURL: subscriptionServiceURL,
	}
}

func (tb *TemplateBuilder) BuildConfirmationEmail(ctx context.Context, email, city, confirmationLink string) (string, error) {
	tb.logger.Infof("Building confirmation email template for email: %s, city: %s", email, city)

	tmpl, err := template.New("confirmation").Parse(ConfirmationEmailTemplate)
	if err != nil {
		tb.logger.Errorf("Failed to parse confirmation template: %v", err)
		return "", err
	}

	data := ConfirmationEmailData{
		City:             city,
		ConfirmationLink: confirmationLink,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		tb.logger.Errorf("Failed to execute confirmation template: %v", err)
		return "", err
	}

	tb.logger.Infof("Confirmation email template built successfully for email: %s, city: %s", email, city)
	return buf.String(), nil
}

func (tb *TemplateBuilder) BuildWeatherUpdateEmail(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int, unsubscribeToken string) (string, error) {
	tb.logger.Infof("Building weather update email template for email: %s, city: %s", email, city)

	tmpl, err := template.New("weather").Parse(WeatherUpdateEmailTemplate)
	if err != nil {
		tb.logger.Errorf("Failed to parse weather template: %v", err)
		return "", err
	}

	unsubscribeLink := ""
	if unsubscribeToken != "" {
		unsubscribeLink = BuildUnsubscribeLink(tb.subscriptionServiceURL, unsubscribeToken)
	}

	data := WeatherUpdateEmailData{
		City:            city,
		Temperature:     strconv.Itoa(temperature),
		Description:     description,
		Humidity:        strconv.Itoa(humidity),
		WindSpeed:       strconv.Itoa(windSpeed),
		UnsubscribeLink: template.HTML(unsubscribeLink),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		tb.logger.Errorf("Failed to execute weather template: %v", err)
		return "", err
	}

	tb.logger.Infof("Weather update email template built successfully for email: %s, city: %s", email, city)
	return buf.String(), nil
}
