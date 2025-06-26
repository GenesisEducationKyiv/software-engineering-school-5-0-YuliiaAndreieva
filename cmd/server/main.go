package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
	"weather-api/internal/adapter/weather/openweathermap"
	"weather-api/internal/adapter/weather/weatherapi"
	"weather-api/internal/util/configutil"
	"weather-api/internal/util/logger"

	"weather-api/internal/adapter/email"
	"weather-api/internal/adapter/repository/postgres"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/service"
	httphandler "weather-api/internal/handler/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := configutil.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	m, err := migrate.New("file://migrations", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Failed to initialize migration: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	fileLogger, err := logger.NewFileLogger("logs", "provider_responses.log")
	if err != nil {
		log.Fatalf("Failed to initialize file logger: %v", err)
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

	httpClient := &http.Client{Timeout: 5 * time.Second}

	weatherAPIProvider := weatherapi.NewClient(weatherapi.ClientOptions{
		APIKey:     cfg.WeatherAPIKey,
		BaseURL:    cfg.WeatherAPIBaseURL,
		HTTPClient: httpClient,
		Logger:     fileLogger,
	})

	openWeatherMapProvider := openweathermap.NewClient(openweathermap.ClientOptions{
		APIKey:     cfg.OpenWeatherMapAPIKey,
		BaseURL:    cfg.OpenWeatherMapBaseURL,
		HTTPClient: httpClient,
		Logger:     fileLogger,
	})

	chainProvider := weather.NewChainWeatherProvider(openWeatherMapProvider, weatherAPIProvider)

	subscriptionRepo := postgres.NewSubscriptionRepo(db)
	cityRepo := postgres.NewCityRepository(db)

	weatherService := service.NewWeatherService(chainProvider)
	tokenService := service.NewTokenService()
	emailService := service.NewEmailService(emailAdapter)

	subscriptionService := service.NewSubscriptionService(
		subscriptionRepo,
		cityRepo,
		chainProvider,
		tokenService,
		emailService,
	)

	weatherUpdateService := service.NewWeatherUpdateService(subscriptionService, weatherService)

	weatherHandler := httphandler.NewWeatherHandler(weatherService)
	subscriptionHandler := httphandler.NewSubscriptionHandler(subscriptionService)

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
		err := schedulerService.SendWeatherUpdates(context.Background(), domain.FrequencyHourly)
		if err != nil {
			return
		}
	})
	if err != nil {
		return
	}

	_, err = cron.AddFunc("0 0 * * *", func() {
		err := schedulerService.SendWeatherUpdates(context.Background(), domain.FrequencyDaily)
		if err != nil {
			return
		}
	})
	if err != nil {
		return
	}

	cron.Start()

	port := strconv.Itoa(cfg.Port)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server running on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
