package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"email/tests/helpers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("test.env")
	if err != nil {
		return
	}
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
	handler       *httphandler.EmailHandler
	router        *gin.Engine
	mailHogClient *helpers.MailHogClient
}

func setupEmailIntegrationTest() *emailIntegrationTestSetup {
	emailSender := email.NewSMTPSender(email.SMTPConfig{
		Host: getEnvWithDefault("SMTP_HOST", "localhost"),
		Port: getEnvAsIntWithDefault("SMTP_PORT", 1025),
		User: getEnvWithDefault("SMTP_USERNAME", ""),
		Pass: getEnvWithDefault("SMTP_PASSWORD", ""),
	}, logger.NewLogrusLogger())

	templateBuilder := email.NewTemplateBuilder(logger.NewLogrusLogger())
	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger.NewLogrusLogger(), "http://localhost:8081")
	handler := httphandler.NewEmailHandler(useCase, logger.NewLogrusLogger())

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/send/confirmation", handler.SendConfirmationEmail)
	router.POST("/send/weather-update", handler.SendWeatherUpdateEmail)

	mailHogHost := getEnvWithDefault("MAILHOG_HOST", "localhost")
	mailHogWebPort := getEnvWithDefault("MAILHOG_WEB_PORT", "8025")
	mailHogURL := fmt.Sprintf("http://%s:%s", mailHogHost, mailHogWebPort)
	mailHogClient := helpers.NewMailHogClient(mailHogURL)

	return &emailIntegrationTestSetup{
		handler:       handler,
		router:        router,
		mailHogClient: mailHogClient,
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
	ts := setupEmailIntegrationTest()

	t.Run("Valid confirmation email request - verify email sent", func(t *testing.T) {
		ts.mailHogClient.ClearMailHog(t)

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

		emailSent := ts.mailHogClient.CheckEmailSent(t, "test@example.com", "Confirm Subscription")
		assert.True(t, emailSent, "Email should have been sent and received by MailHog")
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
	ts := setupEmailIntegrationTest()

	t.Run("Valid weather update email request - verify email sent", func(t *testing.T) {
		ts.mailHogClient.ClearMailHog(t)

		request := domain.WeatherUpdateEmailRequest{
			To:          "weather@example.com",
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

		emailSent := ts.mailHogClient.CheckEmailSent(t, "weather@example.com", "Weather Update")
		assert.True(t, emailSent, "Email should have been sent and received by MailHog")
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

func TestEmailServiceIntegration_EmailContentVerification(t *testing.T) {
	ts := setupEmailIntegrationTest()

	t.Run("Verify confirmation email content", func(t *testing.T) {
		ts.mailHogClient.ClearMailHog(t)

		request := domain.ConfirmationEmailRequest{
			To:               "verify@example.com",
			Subject:          "Test Confirmation",
			City:             "Lviv",
			ConfirmationLink: "http://localhost:8082/confirm/test123",
		}

		w, response := ts.makeConfirmationRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)

		foundEmail := ts.mailHogClient.GetEmailContent(t, "verify@example.com")
		require.NotNil(t, foundEmail, "Email should be found in MailHog")

		assert.Contains(t, foundEmail.Content.Body, "Lviv")
		assert.Contains(t, foundEmail.Content.Body, "http://localhost:8082/confirm/test123")
		assert.Contains(t, foundEmail.Content.Body, "Welcome!")
		assert.Contains(t, foundEmail.Content.Body, "Thank you for subscribing")
	})
}
