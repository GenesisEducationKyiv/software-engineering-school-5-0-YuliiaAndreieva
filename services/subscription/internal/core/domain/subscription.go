package domain

import "errors"

type Frequency string

const (
	Daily   Frequency = "daily"
	Weekly  Frequency = "weekly"
	Monthly Frequency = "monthly"
)

type Subscription struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	City        string `json:"city"`
	Frequency   string `json:"frequency"`
	Token       string `json:"token"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type SubscriptionRequest struct {
	Email     string `json:"email" validate:"required,email"`
	City      string `json:"city" validate:"required"`
	Frequency string `json:"frequency" validate:"required"`
}

type SubscriptionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type ConfirmRequest struct {
	Token string `json:"token" validate:"required"`
}

type ConfirmResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UnsubscribeRequest struct {
	Token string `json:"token" validate:"required"`
}

type UnsubscribeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListSubscriptionsQuery struct {
	Frequency string `json:"frequency" validate:"required"`
	LastID    int    `json:"last_id"`
	PageSize  int    `json:"page_size"`
}

type SubscriptionList struct {
	Subscriptions []Subscription `json:"subscriptions"`
	LastIndex     int            `json:"last_index"`
}

type ConfirmationEmailRequest struct {
	To               string `json:"to"`
	Subject          string `json:"subject"`
	City             string `json:"city"`
	ConfirmationLink string `json:"confirmationLink"`
}

var ErrDuplicateSubscription = errors.New("subscription already exists")

type SubscriptionEvent struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
	Token     string `json:"token"`
}

type ConfirmedEvent struct {
	Email string `json:"email"`
	City  string `json:"city"`
	Token string `json:"token"`
}

type UnsubscribedEvent struct {
	Email string `json:"email"`
	City  string `json:"city"`
	Token string `json:"token"`
}
