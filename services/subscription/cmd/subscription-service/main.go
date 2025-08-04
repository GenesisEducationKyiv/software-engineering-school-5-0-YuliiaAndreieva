package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "proto/subscription"
	"subscription/internal/adapter/database"
	grpchandler "subscription/internal/adapter/grpc"
	httphandler "subscription/internal/adapter/http"
	"subscription/internal/adapter/logger"
	"subscription/internal/config"
	"subscription/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	loggerInstance := logger.NewLogrusLogger()

	var db *gorm.DB
	var err error

	for i := 0; i < cfg.Timeout.DatabaseMaxRetries; i++ {
		db, err = gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
		if err == nil {
			break
		}
		loggerInstance.Warnf("Failed to connect to database, retrying in %v... (attempt %d/%d)", cfg.Timeout.DatabaseRetryDelay, i+1, cfg.Timeout.DatabaseMaxRetries)
		time.Sleep(cfg.Timeout.DatabaseRetryDelay)
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database after %d attempts: %v", cfg.Timeout.DatabaseMaxRetries, err))
	}

	if err := db.AutoMigrate(&database.Subscription{}); err != nil {
		loggerInstance.Errorf("Failed to auto-migrate database: %v", err)
		panic(fmt.Sprintf("Failed to auto-migrate database: %v", err))
	}

	repo := database.NewSubscriptionRepo(db, loggerInstance)

	httpClient := &http.Client{Timeout: cfg.Timeout.HTTPClientTimeout}
	emailClient := httphandler.NewEmailClient(cfg.Email.ServiceURL, httpClient, loggerInstance)
	tokenClient := httphandler.NewTokenClient(cfg.Token.ServiceURL, httpClient, loggerInstance)

	subscribeUseCase := usecase.NewSubscribeUseCase(repo, tokenClient, emailClient, loggerInstance, cfg)
	confirmUseCase := usecase.NewConfirmSubscriptionUseCase(repo, tokenClient, loggerInstance)
	unsubscribeUseCase := usecase.NewUnsubscribeUseCase(repo, tokenClient, loggerInstance)
	listByFrequencyUseCase := usecase.NewListByFrequencyUseCase(repo, loggerInstance)

	subscriptionHandler := httphandler.NewSubscriptionHandler(
		subscribeUseCase,
		confirmUseCase,
		unsubscribeUseCase,
		listByFrequencyUseCase,
		loggerInstance,
	)

	grpcHandler := grpchandler.NewSubscriptionHandler(listByFrequencyUseCase)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "subscription"})
	})

	r.POST("/subscribe", subscriptionHandler.Subscribe)
	r.GET("/confirm/:token", subscriptionHandler.Confirm)
	r.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
	r.POST("/subscriptions/list", subscriptionHandler.ListByFrequency)

	httpSrv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterSubscriptionServiceServer(grpcSrv, grpcHandler)

	go func() {
		loggerInstance.Infof("Starting HTTP server on port %s", cfg.Server.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Failed to start HTTP server: " + err.Error())
		}
	}()

	go func() {
		grpcPort := cfg.Server.GRPCPort
		if grpcPort == "" {
			grpcPort = "9090"
		}
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			panic(fmt.Sprintf("Failed to listen for gRPC: %v", err))
		}

		loggerInstance.Infof("Starting gRPC server on port %s", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			panic("Failed to start gRPC server: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		panic("HTTP server forced to shutdown: " + err.Error())
	}
}
