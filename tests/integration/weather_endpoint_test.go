//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	httphandler "weather-api/internal/adapter/handler/http"
	"weather-api/internal/core/domain"
)

type weatherTestServer struct {
	router   *gin.Engine
	services *TestServices
}

func setupWeatherTestServer(t *testing.T) *weatherTestServer {
	testConfig := GetTestConfig()
	os.Setenv("DB_CONN_STR", testConfig.DBConnStr)
	os.Setenv("WEATHER_API_KEY", testConfig.WeatherAPIKey)
	os.Setenv("SMTP_HOST", testConfig.SMTPHost)
	os.Setenv("SMTP_PORT", testConfig.SMTPPort)
	os.Setenv("SMTP_USER", testConfig.SMTPUser)
	os.Setenv("SMTP_PASS", testConfig.SMTPPass)
	os.Setenv("PORT", testConfig.Port)
	os.Setenv("BASE_URL", testConfig.BaseURL)

	services := SetupTestServices(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	weatherHandler := httphandler.NewWeatherHandler(services.WeatherService)
	api := router.Group("/api")
	{
		api.GET("/weather", weatherHandler.GetWeather)
	}

	return &weatherTestServer{
		router:   router,
		services: services,
	}
}

func (ts *weatherTestServer) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)
	return w
}

func (ts *weatherTestServer) cleanup() {
	if ts.services != nil {
		ts.services.Cleanup()
	}
}
func TestWeatherEndpoint_Integration(t *testing.T) {
	ts := setupWeatherTestServer(t)
	defer ts.cleanup()

	t.Run("GetWeather - Success", func(t *testing.T) {
		w := ts.performRequest("GET", "/api/weather?city=Kyiv", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.Weather
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Description)
	})

	t.Run("GetWeather - Missing City", func(t *testing.T) {
		w := ts.performRequest("GET", "/api/weather", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "city parameter is required", response["error"])
	})

	t.Run("GetWeather - Different Cities", func(t *testing.T) {
		cities := []string{"Lviv", "Chernivtsi", "Ternopil"}

		for _, city := range cities {
			t.Run(city, func(t *testing.T) {
				w := ts.performRequest("GET", fmt.Sprintf("/api/weather?city=%s", city), nil)

				assert.Equal(t, http.StatusOK, w.Code)

				var response domain.Weather
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
			})
		}
	})
}
