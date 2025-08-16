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
	"subscription/internal/core/ports/in"
	"subscription/internal/core/ports/out"
	"subscription/internal/core/usecase"

	"github.com/avast/retry-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDB(cfg *config.Config, logger sharedlogger.Logger) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < cfg.Timeout.DatabaseMaxRetries; i++ {
		db, err = gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
		if err == nil {
			break
		}
		logger.Errorf("Failed to connect to database (attempt %d): %v", i+1, err)
		if i < cfg.Timeout.DatabaseMaxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		logger.Fatalf("Failed to connect to database after %d attempts: %v", cfg.Timeout.DatabaseMaxRetries, err)
	}

	if err := db.AutoMigrate(&database.Subscription{}); err != nil {
		logger.Fatalf("Failed to auto-migrate database: %v", err)
	}

	logger.Infof("Successfully connected to database")
	return db
}

func setupRabbitMQ(cfg *config.Config, logger sharedlogger.Logger) *messaging.RabbitMQPublisher {
	var eventPublisher *messaging.RabbitMQPublisher
	err := retry.Do(
		func() error {
			var err error
			eventPublisher, err = messaging.NewRabbitMQPublisher(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange, logger)
			return err
		},
		retry.Attempts(10),
		retry.Delay(3*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			logger.Errorf("Failed to connect to RabbitMQ (attempt %d): %v", n+1, err)
		}),
	)

	if err != nil {
		logger.Fatalf("Failed to create RabbitMQ publisher after retries: %v", err)
	}

	logger.Infof("Successfully connected to RabbitMQ")
	return eventPublisher
}

func setupUseCases(repo out.SubscriptionRepository, tokenClient out.TokenService, eventPublisher *messaging.RabbitMQPublisher, metricsCollector out.MetricsCollector, logger sharedlogger.Logger, cfg *config.Config) (in.SubscribeUseCase, in.ConfirmSubscriptionUseCase, in.UnsubscribeUseCase, in.ListByFrequencyUseCase) {
	eventPublisherWithMetrics := messaging.NewWithMetrics(eventPublisher, metricsCollector)

	subscribeUseCase := usecase.NewSubscribeUseCase(repo, tokenClient, eventPublisherWithMetrics, logger, cfg)
	confirmUseCase := usecase.NewConfirmSubscriptionUseCase(repo, tokenClient, logger)
	unsubscribeUseCase := usecase.NewUnsubscribeUseCase(repo, tokenClient, logger)
	listByFrequencyUseCase := usecase.NewListByFrequencyUseCase(repo, logger)

	return subscribeUseCase, confirmUseCase, unsubscribeUseCase, listByFrequencyUseCase
}

func setupHandlers(subscribeUseCase in.SubscribeUseCase, confirmUseCase in.ConfirmSubscriptionUseCase, unsubscribeUseCase in.UnsubscribeUseCase, listByFrequencyUseCase in.ListByFrequencyUseCase, metricsCollector out.MetricsCollector, logger sharedlogger.Logger) (*httphandler.SubscriptionHandler, *grpchandler.SubscriptionHandler) {
	subscribeUseCaseWithMetrics := usecase.NewSubscribeWithMetrics(subscribeUseCase, metricsCollector)
	confirmUseCaseWithMetrics := usecase.NewConfirmSubscriptionWithMetrics(confirmUseCase, metricsCollector)
	unsubscribeUseCaseWithMetrics := usecase.NewUnsubscribeWithMetrics(unsubscribeUseCase, metricsCollector)

	subscriptionHandler := httphandler.NewSubscriptionHandler(
		subscribeUseCaseWithMetrics,
		confirmUseCaseWithMetrics,
		unsubscribeUseCaseWithMetrics,
		listByFrequencyUseCase,
		logger,
	)

	grpcHandler := grpchandler.NewSubscriptionHandler(listByFrequencyUseCase)

	return subscriptionHandler, grpcHandler
}

func setupHTTPServer(cfg *config.Config, subscriptionHandler *httphandler.SubscriptionHandler) *http.Server {
	r := gin.Default()

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "subscription"})
	})

	r.POST("/subscribe", subscriptionHandler.Subscribe)
	r.GET("/confirm/:token", subscriptionHandler.Confirm)
	r.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
	r.POST("/subscriptions/list", subscriptionHandler.ListByFrequency)

	return &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}
}

func setupGRPCServer(grpcHandler *grpchandler.SubscriptionHandler) *grpc.Server {
	grpcSrv := grpc.NewServer()
	pb.RegisterSubscriptionServiceServer(grpcSrv, grpcHandler)
	return grpcSrv
}

func startServers(httpSrv *http.Server, grpcSrv *grpc.Server, cfg *config.Config, logger sharedlogger.Logger) {
	go func() {
		logger.Infof("Starting HTTP server on port %s", cfg.Server.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
		if err != nil {
			logger.Fatalf("Failed to listen for gRPC: %v", err)
		}

		logger.Infof("Starting gRPC server on port %s", cfg.Server.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
}

func gracefulShutdown(httpSrv *http.Server, grpcSrv *grpc.Server, eventPublisher *messaging.RabbitMQPublisher, cfg *config.Config, logger sharedlogger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	if err := eventPublisher.Close(); err != nil {
		logger.Errorf("Failed to close event publisher: %v", err)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	db := setupDB(cfg, loggerInstance)
	repo := database.NewSubscriptionRepo(db, loggerInstance)

	metricsCollector := metrics.NewPrometheusCollector()

	httpClient := &http.Client{Timeout: cfg.Timeout.HTTPClientTimeout}
	tokenClient := httphandler.NewTokenClient(cfg.Token.ServiceURL, httpClient, loggerInstance)

	eventPublisher := setupRabbitMQ(cfg, loggerInstance)
	defer func() {
		if err := eventPublisher.Close(); err != nil {
			loggerInstance.Errorf("Failed to close event publisher: %v", err)
		}
	}()

	subscribeUseCase, confirmUseCase, unsubscribeUseCase, listByFrequencyUseCase := setupUseCases(repo, tokenClient, eventPublisher, metricsCollector, loggerInstance, cfg)

	subscriptionHandler, grpcHandler := setupHandlers(subscribeUseCase, confirmUseCase, unsubscribeUseCase, listByFrequencyUseCase, metricsCollector, loggerInstance)

	httpSrv := setupHTTPServer(cfg, subscriptionHandler)
	grpcSrv := setupGRPCServer(grpcHandler)

	startServers(httpSrv, grpcSrv, cfg, loggerInstance)

	gracefulShutdown(httpSrv, grpcSrv, eventPublisher, cfg, loggerInstance)
}
