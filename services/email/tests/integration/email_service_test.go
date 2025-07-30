package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"email/internal/adapter/email"
	httphandler "email/internal/adapter/http"
	"email/internal/adapter/logger"
	"email/internal/core/domain"
	"email/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	godotenv.Load("test.env")
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

type emailIntegrationTestSetup struct {
	handler *httphandler.EmailHandler
	router  *gin.Engine
}

func setupEmailIntegrationTest(t *testing.T) *emailIntegrationTestSetup {
	logger := logger.NewLogrusLogger()

	smtpConfig := email.SMTPConfig{
		Host: getEnvWithDefault("TEST_SMTP_HOST", "localhost"),
		Port: getEnvAsIntWithDefault("TEST_SMTP_PORT", 1025),
		User: getEnvWithDefault("TEST_SMTP_USER", "test@example.com"),
		Pass: getEnvWithDefault("TEST_SMTP_PASS", ""),
	}

	emailSender := email.NewSMTPSender(smtpConfig, logger)
	templateBuilder := email.NewTemplateBuilder(logger)

	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")
	handler := httphandler.NewEmailHandler(useCase, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/send/confirmation", handler.SendConfirmationEmail)
	router.POST("/send/weather-update", handler.SendWeatherUpdateEmail)

	return &emailIntegrationTestSetup{
		handler: handler,
		router:  router,
	}
}

func (eits *emailIntegrationTestSetup) makeConfirmationRequest(t *testing.T, request domain.ConfirmationEmailRequest) (*httptest.ResponseRecorder, *domain.EmailResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/send/confirmation", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	eits.router.ServeHTTP(w, req)

	var response domain.EmailResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func (eits *emailIntegrationTestSetup) makeWeatherUpdateRequest(t *testing.T, request domain.WeatherUpdateEmailRequest) (*httptest.ResponseRecorder, *domain.EmailResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/send/weather-update", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	eits.router.ServeHTTP(w, req)

	var response domain.EmailResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func TestEmailServiceIntegration_SendConfirmationEmail(t *testing.T) {
	ts := setupEmailIntegrationTest(t)

	t.Run("Valid confirmation email request", func(t *testing.T) {
		request := domain.ConfirmationEmailRequest{
			To:               "test@example.com",
			Subject:          "Confirm Subscription",
			City:             "Kyiv",
			ConfirmationLink: "http://localhost:8082/confirm/token123",
		}

		w, response := ts.makeConfirmationRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})

	t.Run("Invalid email format", func(t *testing.T) {
		request := domain.ConfirmationEmailRequest{
			To:               "invalid-email",
			Subject:          "Confirm Subscription",
			City:             "Kyiv",
			ConfirmationLink: "http://localhost:8082/confirm/token123",
		}

		w, response := ts.makeConfirmationRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})

	t.Run("Empty required fields", func(t *testing.T) {
		request := domain.ConfirmationEmailRequest{
			To:               "",
			Subject:          "",
			City:             "",
			ConfirmationLink: "",
		}

		w, response := ts.makeConfirmationRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})
}

func TestEmailServiceIntegration_SendWeatherUpdateEmail(t *testing.T) {
	ts := setupEmailIntegrationTest(t)

	t.Run("Valid weather update email request", func(t *testing.T) {
		request := domain.WeatherUpdateEmailRequest{
			To:          "test@example.com",
			Subject:     "Weather Update",
			Name:        "User",
			City:        "Kyiv",
			Temperature: 15,
			Description: "Partly cloudy",
			Humidity:    65,
			WindSpeed:   12,
		}

		w, response := ts.makeWeatherUpdateRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})

	t.Run("Invalid email format", func(t *testing.T) {
		request := domain.WeatherUpdateEmailRequest{
			To:          "invalid-email",
			Subject:     "Weather Update",
			Name:        "User",
			City:        "Kyiv",
			Temperature: 15,
			Description: "Partly cloudy",
			Humidity:    65,
			WindSpeed:   12,
		}

		w, response := ts.makeWeatherUpdateRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		request := domain.WeatherUpdateEmailRequest{
			To:          "test@example.com",
			Subject:     "",
			Name:        "",
			City:        "",
			Temperature: 15,
			Description: "",
			Humidity:    65,
			WindSpeed:   12,
		}

		w, response := ts.makeWeatherUpdateRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})
}

func TestEmailServiceIntegration_TemplateGeneration(t *testing.T) {
	logger := logger.NewLogrusLogger()
	templateBuilder := email.NewTemplateBuilder(logger)

	t.Run("Confirmation email template", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		city := "Kyiv"
		confirmationLink := "http://localhost:8082/confirm/token123"

		template, err := templateBuilder.BuildConfirmationEmail(ctx, email, city, confirmationLink)

		assert.NoError(t, err)
		assert.NotEmpty(t, template)
		assert.Contains(t, template, city)
		assert.Contains(t, template, confirmationLink)
		assert.Contains(t, template, "Welcome!")
		assert.Contains(t, template, "Thank you for subscribing")
	})

	t.Run("Weather update email template", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		city := "Kyiv"
		description := "Partly cloudy"
		humidity := 65
		windspeed := 12
		temperature := 12

		template, err := templateBuilder.BuildWeatherUpdateEmail(ctx, email, city, description, humidity, windspeed, temperature)
		assert.NoError(t, err)
		assert.NotEmpty(t, template)
		assert.Contains(t, template, city)
		assert.Contains(t, template, description)
		assert.Contains(t, template, "12°C")
		assert.Contains(t, template, "65%")
		assert.Contains(t, template, "12 km/h")
		assert.Contains(t, template, "Weather Update for")
	})
}
