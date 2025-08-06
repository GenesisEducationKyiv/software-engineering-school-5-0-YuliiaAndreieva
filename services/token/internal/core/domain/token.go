package domain

type GenerateTokenRequest struct {
	Subject   string `json:"subject" validate:"required"`
	ExpiresIn string `json:"expires_in"`
}

type GenerateTokenResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type ValidateTokenResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
