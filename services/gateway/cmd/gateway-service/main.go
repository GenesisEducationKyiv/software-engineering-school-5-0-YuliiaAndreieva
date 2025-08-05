package main

import (
	"context"
	"fmt"
	httphandler "gateway/internal/adapter/http"
	"gateway/internal/config"
	"net/http"
	"os"
	"os/signal"
	sharedlogger "shared/logger"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	httpClient := &http.Client{Timeout: cfg.Timeout.HTTPClientTimeout}

	weatherHandler := httphandler.NewWeatherHandler(cfg.WeatherServiceURL, httpClient, loggerInstance)
	subscriptionHandler := httphandler.NewSubscriptionHandler(cfg.SubscriptionServiceURL, httpClient, loggerInstance)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
			loggerInstance.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		loggerInstance.Fatalf("Server forced to shutdown: %v", err)
	}
}
