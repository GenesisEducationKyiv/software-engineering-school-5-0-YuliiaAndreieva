package email

import (
	"context"
	"fmt"
	"strconv"

	"email-service/internal/core/ports/out"
)

type TemplateBuilder struct {
	logger out.Logger
}

func NewTemplateBuilder(logger out.Logger) out.EmailTemplateBuilder {
	return &TemplateBuilder{
		logger: logger,
	}
}

func (tb *TemplateBuilder) BuildConfirmationEmail(ctx context.Context, name, confirmationLink string) (string, error) {
	tb.logger.Infof("Building confirmation email template for %s", name)
	
	template := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome %s!</h2>
			<p>Thank you for registering. Please click the link below to confirm your email:</p>
			<a href="%s">Confirm Email</a>
			<p>If you didn't create this account, please ignore this email.</p>
		</body>
		</html>
	`, name, confirmationLink)
	
	tb.logger.Infof("Confirmation email template built successfully for %s", name)
	return template, nil
}

func (tb *TemplateBuilder) BuildWeatherUpdateEmail(ctx context.Context, name, location, description string, humidity int, windSpeed int, temperature int) (string, error) {
	tb.logger.Infof("Building weather update email template for %s in %s", name, location)
	
	temperatureStr := strconv.Itoa(temperature)
	humidityStr := strconv.Itoa(humidity)
	windSpeedStr := strconv.Itoa(windSpeed)
	
	template := fmt.Sprintf(`
		<html>
		<body>
			<h2>Weather Update for %s</h2>
			<p>Hello %s,</p>
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
	`, location, name, location, temperatureStr, description, humidityStr, windSpeedStr)
	
	tb.logger.Infof("Weather update email template built successfully for %s in %s", name, location)
	return template, nil
}
