package dto

type TokenGenerationResponse struct {
	Token string `json:"token"`
}

type TokenValidationResponse struct {
	Valid bool `json:"valid"`
}
