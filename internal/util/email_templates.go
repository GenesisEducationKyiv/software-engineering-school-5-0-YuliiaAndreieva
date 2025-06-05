package util

import (
	"fmt"
)

func BuildConfirmationEmail(city, token string) (subject, body string) {
	baseURL := GetBaseURL()
	confirmURL := fmt.Sprintf("%s/api/confirm/%s", baseURL, token)
	subject = "Confirm Subscription"
	body = fmt.Sprintf(`
        <html>
            <body>
                <p>Thank you for subscribing to weather updates for %s!</p>
                <p>Please click the link below to confirm your subscription:</p>
                <p><a href="%s" style="color: #0066cc; text-decoration: underline;">Confirm your subscription</a></p>
            </body>
        </html>
    `, city, confirmURL)
	return
}

func BuildWeatherUpdateEmail(city string, temperature float64, humidity int, description, token string) (subject, body string) {
	baseURL := GetBaseURL()
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", baseURL, token)
	subject = "Weather Update"
	body = fmt.Sprintf(`
        <html>
            <body>
                <p>Weather in %s: Temp %.2f°C, Humidity %d%%, %s</p>
                <p><a href="%s" style="color: #0066cc; text-decoration: underline;">Unsubscribe</a></p>
            </body>
        </html>
    `, city, temperature, humidity, description, unsubscribeURL)
	return
}
