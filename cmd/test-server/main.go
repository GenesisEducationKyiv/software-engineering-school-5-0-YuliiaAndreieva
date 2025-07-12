package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"weather-api/internal/adapter/cache/core/redis"
	weathercache "weather-api/internal/adapter/cache/weather"
	"weather-api/internal/adapter/email"
	httphandler "weather-api/internal/adapter/handler/http"
	"weather-api/internal/adapter/repository/postgres"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/adapter/weather/openweathermap"
	"weather-api/internal/adapter/weather/weatherapi"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/service"
	"weather-api/internal/core/usecase"
	"weather-api/internal/util/configutil"
	"weather-api/internal/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/robfig/cron/v3"
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

func main() {
	cfg, err := configutil.LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	m, err := migrate.New("file://migrations", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Unable to initialize migration: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Unable to apply migrations: %v", err)
	}

	fileLogger, err := logger.NewFileLogger("logs", "provider_responses.log")
	if err != nil {
		log.Fatalf("Unable to initialize file logger: %v", err)
	}
	defer func() {
		if closeErr := fileLogger.Close(); closeErr != nil {
			log.Printf("Error closing file logger: %v", closeErr)
		}
	}()

	emailAdapter := email.NewSender(email.SenderOptions{
		Host: cfg.SMTPHost,
		Port: cfg.SMTPPort,
		User: cfg.SMTPUser,
		Pass: cfg.SMTPPass,
	})

	mockClient := &MockHTTPClient{}

	weatherAPIProvider := weatherapi.NewClient(weatherapi.ClientOptions{
		APIKey:     cfg.WeatherAPIKey,
		BaseURL:    cfg.WeatherAPIBaseURL,
		HTTPClient: mockClient,
		Logger:     fileLogger,
	})

	openWeatherMapProvider := openweathermap.NewClient(openweathermap.ClientOptions{
		APIKey:     cfg.OpenWeatherMapAPIKey,
		BaseURL:    cfg.OpenWeatherMapBaseURL,
		HTTPClient: mockClient,
		Logger:     fileLogger,
	})

	chainProvider := weather.NewChainWeatherProvider(weatherAPIProvider, openWeatherMapProvider)

	cache := redis.NewCache(redis.CacheOptions{
		Address:      cfg.RedisAddress,
		TTL:          cfg.RedisTTL,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})
	weatherCache := weathercache.NewCache(cache)

	subscriptionRepo := postgres.NewSubscriptionRepo(db)
	cityRepo := postgres.NewCityRepository(db)

	cachedProvider := weather.NewCachedWeatherProvider(weatherCache, chainProvider)
	weatherService := service.NewWeatherService(cachedProvider)
	tokenService := service.NewTokenService(subscriptionRepo)
	emailService := service.NewEmailService(emailAdapter)
	cityService := service.NewCityService(cityRepo, cachedProvider)

	weatherUseCase := usecase.NewWeatherUseCase(cachedProvider)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, cityRepo, chainProvider, tokenService, emailService)
	subscribeUseCase := usecase.NewSubscribeUseCase(subscriptionRepo, subscriptionService, cityService, tokenService, emailService)
	confirmUseCase := usecase.NewConfirmSubscriptionUseCase(subscriptionRepo, tokenService, emailService)
	unsubscribeUseCase := usecase.NewUnsubscribeUseCase(subscriptionRepo, tokenService)

	weatherUpdateService := service.NewWeatherUpdateService(subscriptionService, weatherService)

	weatherHandler := httphandler.NewWeatherHandler(weatherUseCase)
	subscriptionHandler := httphandler.NewSubscriptionHandler(subscribeUseCase, confirmUseCase, unsubscribeUseCase)

	r := gin.Default()

	r.Static("/web", "./web")

	api := r.Group("/api")
	{
		api.GET("/weather", weatherHandler.GetWeather)
		api.POST("/subscribe", subscriptionHandler.Subscribe)
		api.GET("/confirm/:token", subscriptionHandler.Confirm)
		api.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
	}

	r.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})

	schedulerService := service.NewSchedulerService(weatherUpdateService, emailService)
	cron := cron.New()
	_, err = cron.AddFunc("* * * * *", func() {
		updateErr := schedulerService.SendWeatherUpdates(context.Background(), domain.FrequencyHourly)
		if updateErr != nil {
			log.Printf("Unable to send hourly weather updates: %v", updateErr)
		}
	})
	if err != nil {
		log.Printf("Unable to add hourly cron job: %v", err)
		return
	}

	_, err = cron.AddFunc("0 0 * * *", func() {
		updateErr := schedulerService.SendWeatherUpdates(context.Background(), domain.FrequencyDaily)
		if updateErr != nil {
			log.Printf("Unable to send daily weather updates: %v", updateErr)
		}
	})
	if err != nil {
		log.Printf("Unable to add daily cron job: %v", err)
		return
	}

	cron.Start()

	port := strconv.Itoa(cfg.Port)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}

	log.Printf("Test server running on %s (with mocked Weather API)", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
