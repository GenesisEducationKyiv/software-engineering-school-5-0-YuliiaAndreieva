package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandler "gateway/internal/adapter/http"
	"gateway/internal/adapter/logger"
	"gateway/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	loggerInstance := logger.NewLogger()

	httpClient := &http.Client{Timeout: 30 * time.Second}

	weatherHandler := httphandler.NewWeatherHandler(cfg.WeatherServiceURL, httpClient, loggerInstance)
	subscriptionHandler := httphandler.NewSubscriptionHandler(cfg.SubscriptionServiceURL, httpClient, loggerInstance)

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
	})

	router.GET("/weather", weatherHandler.Get)

	router.POST("/subscribe", subscriptionHandler.Subscribe)
	router.GET("/confirm/:token", subscriptionHandler.Confirm)
	router.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
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
	if err := srv.Shutdown(ctx); err != nil {
		panic("Server forced to shutdown: " + err.Error())
	}
}
