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
	To               string `json:"to"`
	Subject          string `json:"subject"`
	City             string `json:"city"`
	ConfirmationLink string `json:"confirmationLink"`
}

type SubscriptionCreatedEvent struct {
	Email       string `json:"email"`
	City        string `json:"city"`
	Frequency   string `json:"frequency"`
	Token       string `json:"token"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type WeatherUpdateEmailRequest struct {
	To               string  `json:"to" validate:"required,email"`
	Subject          string  `json:"subject" validate:"required"`
	Name             string  `json:"name" validate:"required"`
	City             string  `json:"city" validate:"required"`
	Temperature      float64 `json:"temperature"`
	Description      string  `json:"description" validate:"required"`
	Humidity         float64 `json:"humidity"`
	WindSpeed        float64 `json:"windSpeed"`
	UnsubscribeToken string  `json:"unsubscribeToken"`
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
