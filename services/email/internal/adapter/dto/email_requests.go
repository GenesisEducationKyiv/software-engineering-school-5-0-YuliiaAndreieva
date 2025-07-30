package dto

type ConfirmationEmailRequest struct {
	To               string `json:"to"`
	Subject          string `json:"subject"`
	City             string `json:"city"`
	ConfirmationLink string `json:"confirmationLink"`
}

type WeatherUpdateEmailRequest struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	Name        string `json:"name"`
	City        string `json:"city"`
	Temperature int    `json:"temperature"`
	Description string `json:"description"`
	Humidity    int    `json:"humidity"`
	WindSpeed   int    `json:"windSpeed"`
}
