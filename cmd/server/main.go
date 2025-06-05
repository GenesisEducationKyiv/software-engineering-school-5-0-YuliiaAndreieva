package main

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"strconv"
	"time"
	"weather-api/internal/core/domain"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"weather-api/internal/adapter/email"
	"weather-api/internal/adapter/repository/postgres"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/service"
	httphandler "weather-api/internal/handler/http"
	"weather-api/internal/util"
)

func main() {
	cfg, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	m, err := migrate.New("file://migrations", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Failed to initialize migration: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	emailAdapter := email.NewSMTPEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	weatherAdapter := weather.NewWeatherService(cfg.WeatherAPIKey)
	repo := postgres.NewSubscriptionRepo(db)

	weatherService := service.NewWeatherService(weatherAdapter)
	tokenService := service.NewTokenService()
	subscriptionService := service.NewSubscriptionService(repo, weatherService, emailAdapter, tokenService)
	emailService := service.NewEmailService(repo, weatherAdapter, emailAdapter)

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

	cron := cron.New()
	cron.AddFunc("0 * * * *", func() { emailService.SendUpdates(context.Background(), domain.FrequencyHourly) })
	cron.AddFunc("0 0 * * *", func() { emailService.SendUpdates(context.Background(), domain.FrequencyDaily) })
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
