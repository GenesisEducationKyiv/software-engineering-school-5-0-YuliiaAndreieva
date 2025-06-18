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

func BuildWeatherUpdateEmail(city string, temperature float64, humidity int, description, token string) (subject, body string) {
	baseURL := configutil.GetBaseURL()
	unsubscribeURL := baseURL + "/api/unsubscribe/" + token
	subject = "Weather Update"
	tempStr := strconv.FormatFloat(temperature, 'f', 2, 64)
	humidStr := strconv.Itoa(humidity)

	body = "<html><body>" +
		"<p>Weather in " + city + ": Temp " + tempStr + "°C, Humidity " +
		humidStr + "%, " + description + "</p>" +
		`<p><a href="` + unsubscribeURL +
		`" style="color: #0066cc; text-decoration: underline;">Unsubscribe</a></p>` +
		"</body></html>"

	return
}
