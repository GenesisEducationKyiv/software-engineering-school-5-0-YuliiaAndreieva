package http

import (
	"errors"
	"net/http"
	"weather-api/internal/core/ports/in"

	"github.com/gin-gonic/gin"

	httperrors "weather-api/internal/adapter/handler/http/errors"
	"weather-api/internal/adapter/handler/http/request"
	"weather-api/internal/adapter/handler/http/response"
	"weather-api/internal/core/domain"
)

type WeatherHandler struct {
	weatherUseCase in.WeatherUseCase
}

func NewWeatherHandler(weatherUseCase in.WeatherUseCase) *WeatherHandler {
	return &WeatherHandler{weatherUseCase: weatherUseCase}
}

func (h *WeatherHandler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	cityReq := request.NewCityRequest(city)

	if err := cityReq.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	weather, err := h.weatherUseCase.GetWeather(c, city)
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
