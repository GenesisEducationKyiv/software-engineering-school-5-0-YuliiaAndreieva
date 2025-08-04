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
	httphandler "email/internal/adapter/http"
	"email/internal/adapter/logger"
	"email/internal/config"
	"email/internal/core/usecase"
	pb "proto/email"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.SMTP.User == "" || cfg.SMTP.Pass == "" {
		panic("SMTP_USER and SMTP_PASS environment variables are required")
	}

	loggerInstance := logger.NewLogrusLogger()

	smtpConfig := email.SMTPConfig{
		Host: cfg.SMTP.Host,
		Port: cfg.SMTP.Port,
		User: cfg.SMTP.User,
		Pass: cfg.SMTP.Pass,
	}
	emailSender := email.NewSMTPSender(smtpConfig, loggerInstance)
	templateBuilder := email.NewTemplateBuilder(loggerInstance)

	sendEmailUseCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, loggerInstance, cfg.Server.BaseURL)

	emailHandler := httphandler.NewEmailHandler(sendEmailUseCase, loggerInstance)

	grpcHandler := grpchandler.NewEmailHandler(sendEmailUseCase)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "email"})
	})

	r.POST("/send/confirmation", emailHandler.SendConfirmationEmail)
	r.POST("/send/weather-update", emailHandler.SendWeatherUpdateEmail)

	httpSrv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterEmailServiceServer(grpcSrv, grpcHandler)

	go func() {
		loggerInstance.Infof("Starting HTTP server on port %s", cfg.Server.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Failed to start HTTP server: " + err.Error())
		}
	}()

	go func() {
		grpcPort := cfg.Server.GRPCPort
		if grpcPort == "" {
			grpcPort = "9091"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		panic("HTTP server forced to shutdown: " + err.Error())
	}
}
