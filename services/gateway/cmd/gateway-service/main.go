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

func setupHandlers(cfg *config.Config, httpClient *http.Client, logger sharedlogger.Logger) (*httphandler.WeatherHandler, *httphandler.SubscriptionHandler) {
	weatherHandler := httphandler.NewWeatherHandler(cfg.WeatherServiceURL, httpClient, logger)
	subscriptionHandler := httphandler.NewSubscriptionHandler(cfg.SubscriptionServiceURL, httpClient, logger)
	return weatherHandler, subscriptionHandler
}

func setupHTTPServer(cfg *config.Config, weatherHandler *httphandler.WeatherHandler, subscriptionHandler *httphandler.SubscriptionHandler) *http.Server {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/weather", weatherHandler.Get)

	router.POST("/subscribe", subscriptionHandler.Subscribe)
	router.GET("/confirm/:token", subscriptionHandler.Confirm)
	router.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)

	return &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
}

func startServer(srv *http.Server, logger sharedlogger.Logger) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()
}

func gracefulShutdown(srv *http.Server, cfg *config.Config, logger sharedlogger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	httpClient := &http.Client{Timeout: cfg.Timeout.HTTPClientTimeout}

	weatherHandler, subscriptionHandler := setupHandlers(cfg, httpClient, loggerInstance)

	srv := setupHTTPServer(cfg, weatherHandler, subscriptionHandler)

	startServer(srv, loggerInstance)

	gracefulShutdown(srv, cfg, loggerInstance)
}
