package domain

type EmailType string

const (
	EmailTypeConfirmation  EmailType = "confirmation"
	EmailTypeWeatherUpdate EmailType = "weather_update"
)

type SendEmailRequest struct {
	Type    EmailType              `json:"type" validate:"required"`
	Data    map[string]interface{} `json:"data" validate:"required"`
	To      string                 `json:"to" validate:"required,email"`
	Subject string                 `json:"subject" validate:"required"`
}

type ConfirmationEmailData struct {
	City             string `json:"city" validate:"required"`
	ConfirmationLink string `json:"confirmationLink" validate:"required"`
}

type WeatherUpdateEmailData struct {
	City             string  `json:"city" validate:"required"`
	Description      string  `json:"description" validate:"required"`
	Temperature      float64 `json:"temperature" validate:"required"`
	Humidity         float64 `json:"humidity" validate:"required"`
	WindSpeed        float64 `json:"windSpeed" validate:"required"`
	UnsubscribeToken string  `json:"unsubscribeToken,omitempty"`
}
