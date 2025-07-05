package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	httperrors "weather-api/internal/adapter/handler/http/errors"
	"weather-api/internal/adapter/handler/http/request"
	"weather-api/internal/adapter/handler/http/response"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"
)

type WeatherHandler struct {
	weatherService ports.WeatherService
}

func NewWeatherHandler(weatherService ports.WeatherService) *WeatherHandler {
	return &WeatherHandler{weatherService: weatherService}
}

func (h *WeatherHandler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	cityReq := request.NewCityRequest(city)

	if err := cityReq.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	weather, err := h.weatherService.GetWeather(c, city)
	if err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": httperrors.ErrCityNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	resp := response.WeatherResponse{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}
	c.JSON(http.StatusOK, resp)
}
