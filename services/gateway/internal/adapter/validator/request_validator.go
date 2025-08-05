package validator

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RequestValidator struct {
	validate *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		validate: validator.New(),
	}
}

type WeatherRequest struct {
	City string `form:"city" validate:"required"`
}

type SubscriptionRequest struct {
	Email     string `json:"email" validate:"required,email"`
	City      string `json:"city" validate:"required"`
	Frequency string `json:"frequency"`
}

type TokenRequest struct {
	Token string `uri:"token" validate:"required"`
}

func ValidateAndBind[T any](v *RequestValidator, c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if err := v.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &req, nil
}

func ValidateAndBindJSON[T any](v *RequestValidator, c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid JSON request: %w", err)
	}

	if err := v.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &req, nil
}

func ValidateAndBindURI[T any](v *RequestValidator, c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindUri(&req); err != nil {
		return nil, fmt.Errorf("invalid URI parameters: %w", err)
	}

	if err := v.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &req, nil
}
