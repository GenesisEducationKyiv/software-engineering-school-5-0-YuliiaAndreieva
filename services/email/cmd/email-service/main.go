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

	"email/internal/adapter/email"
	grpchandler "email/internal/adapter/grpc"
	"email/internal/adapter/messaging"
	"email/internal/config"
	"email/internal/core/ports/in"
	"email/internal/core/usecase"
	pb "proto/email"
	sharedlogger "shared/logger"

	"github.com/avast/retry-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func setupEmailComponents(cfg *config.Config, logger sharedlogger.Logger) (in.SendEmailUseCase, *grpchandler.EmailHandler) {
	smtpConfig := email.SMTPConfig{
		Host: cfg.SMTP.Host,
		Port: cfg.SMTP.Port,
		User: cfg.SMTP.User,
		Pass: cfg.SMTP.Pass,
	}
	emailSender := email.NewSMTPSender(smtpConfig, logger)
	templateBuilder := email.NewTemplateBuilder(logger, cfg.Server.BaseURL, cfg.Server.SubscriptionServiceURL)

	sendEmailUseCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, cfg.Server.BaseURL)
	grpcHandler := grpchandler.NewEmailHandler(sendEmailUseCase)

	return sendEmailUseCase, grpcHandler
}

func setupRabbitMQConsumer(cfg *config.Config, sendEmailUseCase in.SendEmailUseCase, logger sharedlogger.Logger) *messaging.RabbitMQConsumer {
	var consumer *messaging.RabbitMQConsumer
	err := retry.Do(
		func() error {
			var err error
			consumer, err = messaging.NewRabbitMQConsumer(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Queue, sendEmailUseCase, logger, cfg.Server.SubscriptionServiceURL)
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
		logger.Fatalf("Failed to create RabbitMQ consumer after retries: %v", err)
	}

	logger.Infof("Successfully connected to RabbitMQ")
	return consumer
}

func setupHTTPServer(cfg *config.Config) *http.Server {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "email"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}
}

func setupGRPCServer(grpcHandler *grpchandler.EmailHandler) *grpc.Server {
	grpcSrv := grpc.NewServer()
	pb.RegisterEmailServiceServer(grpcSrv, grpcHandler)
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

func gracefulShutdown(httpSrv *http.Server, grpcSrv *grpc.Server, consumer *messaging.RabbitMQConsumer, cfg *config.Config, logger sharedlogger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	if err := consumer.Close(); err != nil {
		logger.Errorf("Failed to close RabbitMQ consumer: %v", err)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	sendEmailUseCase, grpcHandler := setupEmailComponents(cfg, loggerInstance)

	consumer := setupRabbitMQConsumer(cfg, sendEmailUseCase, loggerInstance)
	defer func() {
		if err := consumer.Close(); err != nil {
			loggerInstance.Errorf("Failed to close RabbitMQ consumer: %v", err)
		}
	}()

	if err := consumer.Start(context.Background()); err != nil {
		loggerInstance.Fatalf("Failed to start RabbitMQ consumer: %v", err)
	}

	httpSrv := setupHTTPServer(cfg)
	grpcSrv := setupGRPCServer(grpcHandler)

	startServers(httpSrv, grpcSrv, cfg, loggerInstance)

	gracefulShutdown(httpSrv, grpcSrv, consumer, cfg, loggerInstance)
}
