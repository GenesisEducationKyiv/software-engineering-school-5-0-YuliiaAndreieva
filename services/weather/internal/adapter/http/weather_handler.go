package http

import (
	"net/http"
	"strings"

	"weather/internal/core/domain"
	"weather/internal/core/ports/in"
	"weather/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	getWeatherUseCase in.GetWeatherUseCase
	logger            out.Logger
}

func NewWeatherHandler(
	getWeatherUseCase in.GetWeatherUseCase,
	logger out.Logger,
) *WeatherHandler {
	return &WeatherHandler{
		getWeatherUseCase: getWeatherUseCase,
		logger:            logger,
	}
}

func (h *WeatherHandler) GetWeather(c *gin.Context) {
	req, err := h.bindRequest(c)
	if err != nil {
		h.handleBindingError(c, err)
		return
	}

	if err := h.validateRequest(req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	h.logger.Infof("Received weather request for city: %s", req.City)

	result, err := h.getWeatherUseCase.GetWeather(c.Request.Context(), req)
	if err != nil {
		h.handleUseCaseError(c, err, req.City)
		return
	}

	h.handleSuccess(c, result, req.City)
}

func (h *WeatherHandler) bindRequest(c *gin.Context) (domain.WeatherRequest, error) {
	var req domain.WeatherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return domain.WeatherRequest{}, err
	}
	return req, nil
}

func (h *WeatherHandler) validateRequest(req domain.WeatherRequest) error {
	if strings.TrimSpace(req.City) == "" {
		return &ValidationError{Message: "City is required"}
	}
	return nil
}

func (h *WeatherHandler) handleBindingError(c *gin.Context, err error) {
	h.logger.Errorf("Failed to bind JSON for weather request: %v", err)
	c.JSON(http.StatusBadRequest, domain.WeatherResponse{
		Success: false,
		Message: "Invalid request: " + err.Error(),
	})
}

func (h *WeatherHandler) handleValidationError(c *gin.Context, err error) {
	h.logger.Warnf("Validation error: %v", err)
	c.JSON(http.StatusBadRequest, domain.WeatherResponse{
		Success: false,
		Message: err.Error(),
	})
}

func (h *WeatherHandler) handleUseCaseError(c *gin.Context, err error, city string) {
	h.logger.Errorf("Failed to get weather data for city %s: %v", city, err)
	c.JSON(http.StatusInternalServerError, domain.WeatherResponse{
		Success: false,
		Message: "Failed to get weather data: " + err.Error(),
	})
}

func (h *WeatherHandler) handleSuccess(c *gin.Context, result *domain.WeatherResponse, city string) {
	if result.Success {
		h.logger.Infof("Weather data retrieved successfully for city: %s", city)
		c.JSON(http.StatusOK, result)
	} else {
		h.logger.Warnf("Weather request failed for city %s: %s", city, result.Message)
		c.JSON(http.StatusBadRequest, result)
	}
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
