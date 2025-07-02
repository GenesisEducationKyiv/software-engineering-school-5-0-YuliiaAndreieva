package http

import (
	"errors"
	"net/http"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/service"
	httperrors "weather-api/internal/handler/http/errors"
	"weather-api/internal/handler/http/request"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	weatherService service.WeatherService
}

func NewWeatherHandler(weatherService service.WeatherService) *WeatherHandler {
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

	c.JSON(http.StatusOK, weather)
}
