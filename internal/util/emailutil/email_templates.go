package emailutil

import (
	"strconv"
	"weather-api/internal/util/configutil"
)

func BuildConfirmationEmail(city, token string) (subject, body string) {
	baseURL := configutil.GetBaseURL()
	confirmURL := baseURL + "/api/confirm/" + token
	subject = "Confirm Subscription"
	body = "<html><body>" +
		"<p>Thank you for subscribing to weather updates for " + city + "!</p>" +
		"<p>Please click the link below to confirm your subscription:</p>" +
		`<p><a href="` + confirmURL + `" style="color: #0066cc; text-decoration: underline;">` +
		"Confirm your subscription</a></p>" +
		"</body></html>"
	return
}

type WeatherUpdateEmailOptions struct {
	City        string
	Temperature float64
	Humidity    int
	Description string
	Token       string
}

func BuildWeatherUpdateEmail(opts WeatherUpdateEmailOptions) (subject, body string) {
	baseURL := configutil.GetBaseURL()
	unsubscribeURL := baseURL + "/api/unsubscribe/" + opts.Token
	subject = "Weather Update"
	tempStr := strconv.FormatFloat(opts.Temperature, 'f', 2, 64)
	humidStr := strconv.Itoa(opts.Humidity)

	body = "<html><body>" +
		"<p>Weather in " + opts.City + ": Temp " + tempStr + "°C, Humidity " +
		humidStr + "%, " + opts.Description + "</p>" +
		`<p><a href="` + unsubscribeURL +
		`" style="color: #0066cc; text-decoration: underline;">Unsubscribe</a></p>` +
		"</body></html>"

	return
}
