//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	httphandler "weather-api/internal/adapter/handler/http"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type subscriptionTestServer struct {
	router   *gin.Engine
	services *TestServices
}

func setupSubscriptionTestServer(t *testing.T) *subscriptionTestServer {
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

	subscriptionHandler := httphandler.NewSubscriptionHandler(services.SubscriptionService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		api.POST("/subscribe", subscriptionHandler.Subscribe)
		api.GET("/confirm/:token", subscriptionHandler.Confirm)
		api.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
	}

	return &subscriptionTestServer{
		router:   router,
		services: services,
	}
}

func (ts *subscriptionTestServer) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (ts *subscriptionTestServer) cleanup() {
	if ts.services != nil {
		ts.services.Cleanup()
	}
}

func TestSubscribeEndpoint_Integration(t *testing.T) {
	ts := setupSubscriptionTestServer(t)
	defer ts.cleanup()

	t.Run("Subscribe - Success", func(t *testing.T) {
		subscribeReq := map[string]interface{}{
			"email":     "test@example.com",
			"city":      "Kyiv",
			"frequency": "daily",
		}

		w := ts.performRequest("POST", "/api/subscribe", subscribeReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Subscription successful. Confirmation email sent.", response["message"])
	})

	t.Run("Subscribe - Invalid Frequency", func(t *testing.T) {
		subscribeReq := map[string]interface{}{
			"email":     "test2@example.com",
			"city":      "Kyiv",
			"frequency": "weekly",
		}

		w := ts.performRequest("POST", "/api/subscribe", subscribeReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Subscribe - Duplicate Subscription", func(t *testing.T) {
		subscribeReq := map[string]interface{}{
			"email":     "duplicate@example.com",
			"city":      "Kyiv",
			"frequency": "daily",
		}

		w1 := ts.performRequest("POST", "/api/subscribe", subscribeReq)
		assert.Equal(t, http.StatusOK, w1.Code)

		w2 := ts.performRequest("POST", "/api/subscribe", subscribeReq)
		assert.Equal(t, http.StatusConflict, w2.Code)
	})
}

func TestConfirmEndpoint_Integration(t *testing.T) {
	ts := setupSubscriptionTestServer(t)
	defer ts.cleanup()

	t.Run("Confirm - Invalid Token", func(t *testing.T) {
		w := ts.performRequest("GET", "/api/confirm/invalid-token", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Confirm - Valid Token", func(t *testing.T) {
		subscribeReq := map[string]interface{}{
			"email":     "confirm@example.com",
			"city":      "Kyiv",
			"frequency": "daily",
		}

		w := ts.performRequest("POST", "/api/subscribe", subscribeReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var token string
		err := ts.services.DB.QueryRowContext(context.Background(),
			"SELECT token FROM subscriptions WHERE email = $1", "confirm@example.com").Scan(&token)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		w = ts.performRequest("GET", fmt.Sprintf("/api/confirm/%s", token), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var isConfirmed bool
		err = ts.services.DB.QueryRowContext(context.Background(),
			"SELECT is_confirmed FROM subscriptions WHERE token = $1", token).Scan(&isConfirmed)
		require.NoError(t, err)
		assert.True(t, isConfirmed)
	})
}

func TestUnsubscribeEndpoint_Integration(t *testing.T) {
	ts := setupSubscriptionTestServer(t)
	defer ts.cleanup()

	t.Run("Unsubscribe - Invalid Token", func(t *testing.T) {
		w := ts.performRequest("GET", "/api/unsubscribe/invalid-token", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Unsubscribe - Valid Token", func(t *testing.T) {
		subscribeReq := map[string]interface{}{
			"email":     "unsubscribe@example.com",
			"city":      "Kyiv",
			"frequency": "daily",
		}

		w := ts.performRequest("POST", "/api/subscribe", subscribeReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var token string
		err := ts.services.DB.QueryRowContext(context.Background(),
			"SELECT token FROM subscriptions WHERE email = $1", "unsubscribe@example.com").Scan(&token)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		w = ts.performRequest("GET", fmt.Sprintf("/api/unsubscribe/%s", token), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var count int
		err = ts.services.DB.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM subscriptions WHERE token = $1", token).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestSubscriptionFlow_Integration(t *testing.T) {
	ts := setupSubscriptionTestServer(t)
	defer ts.cleanup()

	t.Run("Complete Subscription Flow", func(t *testing.T) {
		subscribeReq := map[string]interface{}{
			"email":     "flow@example.com",
			"city":      "Kyiv",
			"frequency": "hourly",
		}

		w := ts.performRequest("POST", "/api/subscribe", subscribeReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var token string
		err := ts.services.DB.QueryRowContext(context.Background(),
			"SELECT token FROM subscriptions WHERE email = $1", "flow@example.com").Scan(&token)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		w = ts.performRequest("GET", fmt.Sprintf("/api/confirm/%s", token), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var isConfirmed bool
		err = ts.services.DB.QueryRowContext(context.Background(),
			"SELECT is_confirmed FROM subscriptions WHERE token = $1", token).Scan(&isConfirmed)
		require.NoError(t, err)
		assert.True(t, isConfirmed)

		w = ts.performRequest("GET", fmt.Sprintf("/api/unsubscribe/%s", token), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var count int
		err = ts.services.DB.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM subscriptions WHERE token = $1", token).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
