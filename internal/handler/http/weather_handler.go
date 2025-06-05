package http

import (
	"errors"
	"net/http"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/service"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	weatherService *service.WeatherService
}

func NewWeatherHandler(weatherService *service.WeatherService) *WeatherHandler {
	return &WeatherHandler{weatherService: weatherService}
}

func (h *WeatherHandler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City parameter is required"})
		return
	}
	weather, err := h.weatherService.GetWeather(city)
	if err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "City not found"})
			return
		}
	}
	c.JSON(http.StatusOK, weather)
}
