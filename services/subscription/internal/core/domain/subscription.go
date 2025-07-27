package domain

type SubscriptionRequest struct {
	Email     string `json:"email" validate:"required,email"`
	City      string `json:"city" validate:"required"`
	Frequency string `json:"frequency" validate:"required"`
}

type SubscriptionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ConfirmRequest struct {
	Token string `json:"token" validate:"required"`
}

type ConfirmResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type UnsubscribeRequest struct {
	Token string `json:"token" validate:"required"`
}

type UnsubscribeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type Subscription struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	City        string `json:"city"`
	Frequency   string `json:"frequency"`
	Token       string `json:"token"`
	IsConfirmed bool   `json:"is_confirmed"`
} 