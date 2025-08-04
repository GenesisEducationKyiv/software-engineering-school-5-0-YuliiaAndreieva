package domain

type EmailRequest struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

type EmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ConfirmationEmailRequest struct {
	To               string `json:"to" validate:"required,email"`
	Subject          string `json:"subject" validate:"required"`
	City             string `json:"city" validate:"required"`
	ConfirmationLink string `json:"confirmationLink" validate:"required"`
}

type WeatherUpdateEmailRequest struct {
	To               string `json:"to" validate:"required,email"`
	Subject          string `json:"subject" validate:"required"`
	Name             string `json:"name" validate:"required"`
	City             string `json:"city" validate:"required"`
	Temperature      int    `json:"temperature"`
	Description      string `json:"description" validate:"required"`
	Humidity         int    `json:"humidity"`
	WindSpeed        int    `json:"windSpeed"`
	UnsubscribeToken string `json:"unsubscribeToken"`
}

type EmailDeliveryStatus string

const (
	StatusFailed    EmailDeliveryStatus = "failed"
	StatusDelivered EmailDeliveryStatus = "delivered"
)

type EmailDeliveryResult struct {
	EmailID string
	To      string
	Status  EmailDeliveryStatus
	Error   string
	SentAt  int64
}

type EmailBuilderRequest struct {
	Type    string                 `json:"type" validate:"required"`
	Data    map[string]interface{} `json:"data" validate:"required"`
	BaseURL string                 `json:"base_url"`
}

type EmailBuilderResponse struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Error   string `json:"error,omitempty"`
}
