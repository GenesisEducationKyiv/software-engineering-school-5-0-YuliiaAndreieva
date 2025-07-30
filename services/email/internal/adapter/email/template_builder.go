package email

import (
	"context"
	"fmt"
	"strconv"

	"email/internal/core/ports/out"
)

type TemplateBuilder struct {
	logger out.Logger
}

func NewTemplateBuilder(logger out.Logger) out.EmailTemplateBuilder {
	return &TemplateBuilder{
		logger: logger,
	}
}

func (tb *TemplateBuilder) BuildConfirmationEmail(ctx context.Context, email, city, confirmationLink string) (string, error) {
	tb.logger.Infof("Building confirmation email template for email: %s, city: %s", email, city)

	template := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome!</h2>
			<p>Thank you for subscribing to weather updates for %s. Please click the link below to confirm your email:</p>
			<a href="%s">Confirm Email</a>
			<p>If you didn't create this subscription, please ignore this email.</p>
		</body>
		</html>
	`, city, confirmationLink)

	tb.logger.Infof("Confirmation email template built successfully for email: %s, city: %s", email, city)
	return template, nil
}

func (tb *TemplateBuilder) BuildWeatherUpdateEmail(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int) (string, error) {
	tb.logger.Infof("Building weather update email template for email: %s, city: %s", email, city)

	temperatureStr := strconv.Itoa(temperature)
	humidityStr := strconv.Itoa(humidity)
	windSpeedStr := strconv.Itoa(windSpeed)

	template := fmt.Sprintf(`
		<html>
		<body>
			<h2>Weather Update for %s</h2>
			<p>Hello,</p>
			<p>Here's your weather update for %s:</p>
			<ul>
				<li>Temperature: %s°C</li>
				<li>Description: %s</li>
				<li>Humidity: %s%%</li>
				<li>Wind Speed: %s km/h</li>
			</ul>
			<p>Stay safe and enjoy your day!</p>
		</body>
		</html>
	`, city, city, temperatureStr, description, humidityStr, windSpeedStr)

	tb.logger.Infof("Weather update email template built successfully for email: %s, city: %s", email, city)
	return template, nil
}
