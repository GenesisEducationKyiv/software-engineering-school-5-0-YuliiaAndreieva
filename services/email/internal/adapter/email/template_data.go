package email

import "html/template"

type ConfirmationEmailData struct {
	City             string
	ConfirmationLink string
}

type WeatherUpdateEmailData struct {
	City            string
	Temperature     string
	Description     string
	Humidity        string
	WindSpeed       string
	UnsubscribeLink template.HTML
}
