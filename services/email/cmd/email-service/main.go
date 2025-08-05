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
	"email/internal/core/usecase"
	pb "proto/email"
	sharedlogger "shared/logger"

	"github.com/avast/retry-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.SMTP.User == "" || cfg.SMTP.Pass == "" {
		fmt.Printf("SMTP_USER and SMTP_PASS environment variables are required\n")
		os.Exit(1)
	}

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	smtpConfig := email.SMTPConfig{
		Host: cfg.SMTP.Host,
		Port: cfg.SMTP.Port,
		User: cfg.SMTP.User,
		Pass: cfg.SMTP.Pass,
	}
	emailSender := email.NewSMTPSender(smtpConfig, loggerInstance)
	templateBuilder := email.NewTemplateBuilder(loggerInstance, cfg.Server.BaseURL)

	sendEmailUseCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, loggerInstance, cfg.Server.BaseURL)

	grpcHandler := grpchandler.NewEmailHandler(sendEmailUseCase)

	var consumer *messaging.RabbitMQConsumer
	err := retry.Do(
		func() error {
			var err error
			consumer, err = messaging.NewRabbitMQConsumer(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Queue, sendEmailUseCase, loggerInstance, cfg.Server.BaseURL)
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
		loggerInstance.Fatalf("Failed to create RabbitMQ consumer after retries: %v", err)
	}
	defer consumer.Close()

	loggerInstance.Infof("Successfully connected to RabbitMQ")

	if err := consumer.Start(context.Background()); err != nil {
		loggerInstance.Fatalf("Failed to start RabbitMQ consumer: %v", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "email"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	httpSrv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterEmailServiceServer(grpcSrv, grpcHandler)

	go func() {
		loggerInstance.Infof("Starting HTTP server on port %s", cfg.Server.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInstance.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	go func() {
		grpcPort := cfg.Server.GRPCPort
		if grpcPort == "" {
			grpcPort = "9091"
		}
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			loggerInstance.Fatalf("Failed to listen for gRPC: %v", err)
		}

		loggerInstance.Infof("Starting gRPC server on port %s", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			loggerInstance.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		loggerInstance.Fatalf("HTTP server forced to shutdown: %v", err)
	}
}
