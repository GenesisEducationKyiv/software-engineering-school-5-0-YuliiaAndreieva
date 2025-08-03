package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	httphandler "weather-broadcast/internal/adapter/http"
	"weather-broadcast/internal/adapter/logger"
	"weather-broadcast/internal/config"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	loggerInstance := logger.NewLogrusLogger()

	httpClient := &http.Client{Timeout: cfg.HTTPClientTimeout}

	subscriptionClient := httphandler.NewSubscriptionClient(
		cfg.SubscriptionServiceURL,
		httpClient,
		loggerInstance,
	)

	weatherClient := httphandler.NewWeatherClient(
		cfg.WeatherServiceURL,
		httpClient,
		loggerInstance,
	)

	emailClient := httphandler.NewEmailClient(
		cfg.EmailServiceURL,
		httpClient,
		loggerInstance,
	)

	broadcastUseCase := usecase.NewBroadcastUseCase(
		subscriptionClient,
		weatherClient,
		emailClient,
		loggerInstance,
	)

	broadcastHandler := httphandler.NewBroadcastHandler(broadcastUseCase, loggerInstance)

	c := cron.New(cron.WithLocation(time.UTC))

	c.AddFunc("* * * * *", func() {
		loggerInstance.Infof("Starting hourly weather broadcast")
		if err := broadcastUseCase.Broadcast(context.Background(), domain.Daily); err != nil {
			loggerInstance.Errorf("Hourly broadcast failed: %v", err)
		}
	})

	c.AddFunc("0 8 * * *", func() {
		loggerInstance.Infof("Starting daily weather broadcast")
		if err := broadcastUseCase.Broadcast(context.Background(), domain.Daily); err != nil {
			loggerInstance.Errorf("Daily broadcast failed: %v", err)
		}
	})

	c.Start()
	loggerInstance.Infof("Cron jobs started")

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "weather-broadcast"})
	})

	r.POST("/broadcast", broadcastHandler.Broadcast)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Failed to start server: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.Stop()
	if err := srv.Shutdown(ctx); err != nil {
		panic("Server forced to shutdown: " + err.Error())
	}
}
