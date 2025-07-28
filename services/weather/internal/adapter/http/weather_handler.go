package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"weather-service/internal/core/domain"
	"weather-service/internal/core/ports/in"
)

type WeatherHandler struct {
	getWeatherUseCase in.GetWeatherUseCase
}

func NewWeatherHandler(
	getWeatherUseCase in.GetWeatherUseCase,
) *WeatherHandler {
	return &WeatherHandler{
		getWeatherUseCase: getWeatherUseCase,
	}
}

func (h *WeatherHandler) GetWeather(c *gin.Context) {
	var req domain.WeatherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.WeatherResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.getWeatherUseCase.GetWeather(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.WeatherResponse{
			Success: false,
			Message: "Failed to get weather data: " + err.Error(),
		})
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusBadRequest, result)
	}
}
