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
	sharedlogger "shared/logger"
	"subscription/internal/adapter/database"
	grpchandler "subscription/internal/adapter/grpc"
	httphandler "subscription/internal/adapter/http"
	"subscription/internal/adapter/messaging"
	"subscription/internal/adapter/metrics"
	"subscription/internal/config"
	"subscription/internal/core/usecase"

	"github.com/avast/retry-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		loggerInstance.Errorf("Failed to connect to database: %v", err)
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	if err := db.AutoMigrate(&database.Subscription{}); err != nil {
		loggerInstance.Errorf("Failed to auto-migrate database: %v", err)
		panic(fmt.Sprintf("Failed to auto-migrate database: %v", err))
	}

	repo := database.NewSubscriptionRepo(db, loggerInstance)

	metricsCollector := metrics.NewPrometheusCollector()

	httpClient := &http.Client{Timeout: cfg.Timeout.HTTPClientTimeout}
	emailClient := httphandler.NewEmailClient(cfg.Email.ServiceURL, httpClient, loggerInstance)
	tokenClient := httphandler.NewTokenClient(cfg.Token.ServiceURL, httpClient, loggerInstance)

	var eventPublisher *messaging.RabbitMQPublisher
	err = retry.Do(
		func() error {
			var err error
			eventPublisher, err = messaging.NewRabbitMQPublisher(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange, loggerInstance)
			return err
		},
		retry.Attempts(10),
		retry.Delay(3*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			loggerInstance.Errorf("Failed to connect to RabbitMQ (attempt %d): %v", n+1, err)
		}),
	)

	if err != nil {
		panic(fmt.Sprintf("Failed to create RabbitMQ publisher after retries: %v", err))
	}
	defer eventPublisher.Close()

	loggerInstance.Infof("Successfully connected to RabbitMQ")

	eventPublisherWithMetrics := messaging.NewRabbitMQMetricsPublisherDecorator(eventPublisher, metricsCollector)

	subscribeUseCase := usecase.NewSubscribeUseCase(repo, tokenClient, emailClient, eventPublisherWithMetrics, loggerInstance, cfg)
	confirmUseCase := usecase.NewConfirmSubscriptionUseCase(repo, tokenClient, loggerInstance)
	unsubscribeUseCase := usecase.NewUnsubscribeUseCase(repo, tokenClient, loggerInstance)
	listByFrequencyUseCase := usecase.NewListByFrequencyUseCase(repo, loggerInstance)

	subscribeUseCaseWithMetrics := usecase.NewSubscribeMetricsDecorator(subscribeUseCase, metricsCollector)
	confirmUseCaseWithMetrics := usecase.NewConfirmSubscriptionMetricsDecorator(confirmUseCase, metricsCollector)
	unsubscribeUseCaseWithMetrics := usecase.NewUnsubscribeMetricsDecorator(unsubscribeUseCase, metricsCollector)

	subscriptionHandler := httphandler.NewSubscriptionHandler(
		subscribeUseCaseWithMetrics,
		confirmUseCaseWithMetrics,
		unsubscribeUseCaseWithMetrics,
		listByFrequencyUseCase,
		loggerInstance,
	)

	grpcHandler := grpchandler.NewSubscriptionHandler(listByFrequencyUseCase)

	r := gin.Default()

	metricsMiddleware := httphandler.NewMetricsMiddleware(metricsCollector)
	r.Use(metricsMiddleware.MetricsMiddleware())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
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
			grpcPort = "9093"
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
