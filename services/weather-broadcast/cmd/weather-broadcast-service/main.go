package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	sharedlogger "shared/logger"
	"syscall"
	"time"
	grpcclient "weather-broadcast/internal/adapter/grpc"
	httphandler "weather-broadcast/internal/adapter/http"
	"weather-broadcast/internal/config"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/internal/core/ports/in"
	"weather-broadcast/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
)

func setupGRPCClients(cfg *config.Config, logger sharedlogger.Logger) (*grpcclient.SubscriptionClient, *grpcclient.EmailClient, *grpcclient.WeatherClient) {
	subscriptionClient, err := grpcclient.NewSubscriptionClient(cfg.SubscriptionGRPCURL, logger)
	if err != nil {
		logger.Fatalf("Failed to create subscription gRPC client: %v", err)
	}

	emailClient, err := grpcclient.NewEmailClient(cfg.EmailGRPCURL, logger)
	if err != nil {
		logger.Fatalf("Failed to create email gRPC client: %v", err)
	}

	weatherClient, err := grpcclient.NewWeatherClient(cfg.WeatherGRPCURL, logger)
	if err != nil {
		logger.Fatalf("Failed to create weather gRPC client: %v", err)
	}

	return subscriptionClient, emailClient, weatherClient
}

func setupUseCase(subscriptionClient *grpcclient.SubscriptionClient, weatherClient *grpcclient.WeatherClient, emailClient *grpcclient.EmailClient, logger sharedlogger.Logger) in.BroadcastUseCase {
	return usecase.NewBroadcastUseCase(subscriptionClient, weatherClient, emailClient, logger)
}

func setupHandler(broadcastUseCase in.BroadcastUseCase, logger sharedlogger.Logger) *httphandler.BroadcastHandler {
	return httphandler.NewBroadcastHandler(broadcastUseCase, logger)
}

func setupCronJobs(broadcastUseCase in.BroadcastUseCase, logger sharedlogger.Logger) *cron.Cron {
	c := cron.New(cron.WithLocation(time.UTC))

	c.AddFunc("* * * * *", func() {
		logger.Infof("Starting hourly weather broadcast")
		if err := broadcastUseCase.Broadcast(context.Background(), domain.Daily); err != nil {
			logger.Errorf("Hourly broadcast failed: %v", err)
		}
	})

	c.AddFunc("0 8 * * *", func() {
		logger.Infof("Starting daily weather broadcast")
		if err := broadcastUseCase.Broadcast(context.Background(), domain.Daily); err != nil {
			logger.Errorf("Daily broadcast failed: %v", err)
		}
	})

	c.Start()
	logger.Infof("Cron jobs started")
	return c
}

func setupHTTPServer(cfg *config.Config, broadcastHandler *httphandler.BroadcastHandler) *http.Server {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "weather-broadcast"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/broadcast", broadcastHandler.Broadcast)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Infof("Server stopped")
}

func validateConfig(cfg *config.Config, logger sharedlogger.Logger) {
	if cfg.SubscriptionServiceURL == "" {
		logger.Fatalf("SUBSCRIPTION_SERVICE_URL environment variable is required")
	}
	if cfg.WeatherServiceURL == "" {
		logger.Fatalf("WEATHER_SERVICE_URL environment variable is required")
	}
	if cfg.EmailServiceURL == "" {
		logger.Fatalf("EMAIL_SERVICE_URL environment variable is required")
	}
	if cfg.SubscriptionGRPCURL == "" {
		logger.Fatalf("SUBSCRIPTION_GRPC_URL environment variable is required")
	}
	if cfg.EmailGRPCURL == "" {
		logger.Fatalf("EMAIL_GRPC_URL environment variable is required")
	}
	if cfg.WeatherGRPCURL == "" {
		logger.Fatalf("WEATHER_GRPC_URL environment variable is required")
	}
	if cfg.Port == 0 {
		logger.Fatalf("PORT environment variable is required")
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.LogInitial, cfg.LogThereafter, cfg.LogTick)

	validateConfig(cfg, loggerInstance)

	subscriptionClient, emailClient, weatherClient := setupGRPCClients(cfg, loggerInstance)

	broadcastUseCase := setupUseCase(subscriptionClient, weatherClient, emailClient, loggerInstance)
	broadcastHandler := setupHandler(broadcastUseCase, loggerInstance)

	setupCronJobs(broadcastUseCase, loggerInstance)

	srv := setupHTTPServer(cfg, broadcastHandler)

	startServer(srv, loggerInstance)

	gracefulShutdown(srv, cfg, loggerInstance)
}
