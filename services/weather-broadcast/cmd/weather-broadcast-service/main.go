package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	grpcclient "weather-broadcast/internal/adapter/grpc"
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

	subscriptionClient, err := grpcclient.NewSubscriptionClient(cfg.SubscriptionGRPCURL, loggerInstance)
	if err != nil {
		panic(fmt.Sprintf("Failed to create subscription gRPC client: %v", err))
	}

	emailClient, err := grpcclient.NewEmailClient(cfg.EmailGRPCURL, loggerInstance)
	if err != nil {
		panic(fmt.Sprintf("Failed to create email gRPC client: %v", err))
	}

	weatherClient, err := grpcclient.NewWeatherClient(cfg.WeatherGRPCURL, loggerInstance)
	if err != nil {
		panic(fmt.Sprintf("Failed to create weather gRPC client: %v", err))
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	c.Stop()
	if err := srv.Shutdown(ctx); err != nil {
		panic("Server forced to shutdown: " + err.Error())
	}
}
