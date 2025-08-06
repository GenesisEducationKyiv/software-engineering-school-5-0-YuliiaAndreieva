package email

const (
	ConfirmationEmailTemplate = `
		<html>
		<body>
			<h2>Welcome!</h2>
			<p>Thank you for subscribing to weather updates for {{.City}}. Please click the link below to confirm your email:</p>
			<a href="{{.ConfirmationLink}}">Confirm Email</a>
			<p>If you didn't create this subscription, please ignore this email.</p>
		</body>
		</html>
	`

	WeatherUpdateEmailTemplate = `
		<html>
		<body>
			<h2>Weather Update for {{.City}}</h2>
			<p>Hello,</p>
			<p>Here's your weather update for {{.City}}:</p>
			<ul>
				<li>Temperature: {{.Temperature}}°C</li>
				<li>Description: {{.Description}}</li>
				<li>Humidity: {{.Humidity}}%</li>
				<li>Wind Speed: {{.WindSpeed}} km/h</li>
			</ul>
			<p>Stay safe and enjoy your day!</p>
			{{if .UnsubscribeLink}}
			{{.UnsubscribeLink}}
			{{end}}
		</body>
		</html>
	`
)

func BuildUnsubscribeLink(subscriptionServiceURL, token string) string {
	return `<p><a href="` + subscriptionServiceURL + `/unsubscribe/` + token + `">Unsubscribe from weather updates</a></p>`
}
