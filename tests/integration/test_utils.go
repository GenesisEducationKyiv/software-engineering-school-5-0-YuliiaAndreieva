//go:build integration
// +build integration

package integration

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"weather-api/internal/adapter/email"
	"weather-api/internal/adapter/repository/postgres"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/service"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

type MockHTTPClient struct{}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	var cityName string
	if strings.Contains(url, "q=") {
		parts := strings.Split(url, "q=")
		if len(parts) > 1 {
			cityName = strings.Split(parts[1], "&")[0]
		}
	}

	if cityName == "" {
		cityName = "Kyiv"
	}

	weatherResponse := fmt.Sprintf(`{
		"location": {
			"name": "%s",
			"region": "Test Region",
			"country": "Ukraine",
			"lat": 50.45,
			"lon": 30.52,
			"localtime": "2024-01-01 12:00"
		},
		"current": {
			"temp_c": 20.5,
			"humidity": 60,
			"condition": {
				"text": "Sunny"
			}
		}
	}`, cityName)

	searchResponse := fmt.Sprintf(`[
		{
			"id": 1,
			"name": "%s",
			"region": "Test Region",
			"country": "Ukraine",
			"lat": 50.45,
			"lon": 30.52
		}
	]`, cityName)

	var responseBody string
	if strings.Contains(req.URL.Path, "/search.json") {
		responseBody = searchResponse
	} else {
		responseBody = weatherResponse
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
		Header:     make(http.Header),
	}, nil
}

type TestConfig struct {
	DBConnStr     string
	WeatherAPIKey string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	Port          string
	BaseURL       string
}

func GetTestConfig() TestConfig {
	return TestConfig{
		DBConnStr:     "postgres://test:test@localhost:5433/weather_test?sslmode=disable",
		WeatherAPIKey: "test-api-key",
		SMTPHost:      "localhost",
		SMTPPort:      "1025",
		SMTPUser:      "test@example.com",
		SMTPPass:      "test-password",
		Port:          "8080",
		BaseURL:       "http://localhost:8080",
	}
}

func SetupTestDatabase(t *testing.T) *sql.DB {
	testConfig := GetTestConfig()

	db, err := sql.Open("postgres", testConfig.DBConnStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	m, err := migrate.New("file://../../migrations", testConfig.DBConnStr)
	require.NoError(t, err)

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	return db
}

func CleanupTestDatabase(t *testing.T, db *sql.DB) {
	tables := []string{"subscriptions", "cities"}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to clean table %s: %v", table, err)
		}
	}

	_, err := db.Exec("ALTER SEQUENCE cities_id_seq RESTART WITH 1")
	if err != nil {
		t.Logf("Warning: failed to reset cities sequence: %v", err)
	}

	_, err = db.Exec("ALTER SEQUENCE subscriptions_id_seq RESTART WITH 1")
	if err != nil {
		t.Logf("Warning: failed to reset subscriptions sequence: %v", err)
	}
}

func CreateTestDatabase(t *testing.T) (*sql.DB, func()) {
	db := SetupTestDatabase(t)

	cleanup := func() {
		CleanupTestDatabase(t, db)
		db.Close()
	}

	return db, cleanup
}

type TestServices struct {
	DB                   *sql.DB
	SubscriptionService  *service.SubscriptionService
	WeatherService       service.WeatherService
	WeatherUpdateService *service.WeatherUpdateService
	EmailService         *service.EmailService
	Cleanup              func()
}

func SetupTestServices(t *testing.T) *TestServices {
	testConfig := GetTestConfig()

	db, cleanup := CreateTestDatabase(t)

	smtpPort, err := strconv.Atoi(testConfig.SMTPPort)
	require.NoError(t, err)

	emailAdapter := email.NewEmailSender(testConfig.SMTPHost, smtpPort, testConfig.SMTPUser, testConfig.SMTPPass)
	weatherAdapter := weather.NewWeatherAPIClient(
		testConfig.WeatherAPIKey,
		"http://api.weatherapi.com/v1",
		&MockHTTPClient{},
	)

	subscriptionRepo := postgres.NewSubscriptionRepo(db)
	cityRepo := postgres.NewCityRepository(db)

	weatherService := service.NewWeatherService(weatherAdapter)
	tokenService := service.NewTokenService()
	emailService := service.NewEmailService(emailAdapter)

	subscriptionService := service.NewSubscriptionService(
		subscriptionRepo,
		cityRepo,
		weatherAdapter,
		tokenService,
		emailService,
	)

	weatherUpdateService := service.NewWeatherUpdateService(subscriptionService, weatherService)

	return &TestServices{
		DB:                   db,
		SubscriptionService:  subscriptionService,
		WeatherService:       weatherService,
		WeatherUpdateService: weatherUpdateService,
		EmailService:         emailService,
		Cleanup:              cleanup,
	}
}
