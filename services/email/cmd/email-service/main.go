package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email-service/internal/adapter/email"
	"email-service/internal/adapter/logger"
	httphandler "email-service/internal/adapter/http"
	"email-service/internal/config"
	"email-service/internal/core/usecase"

	"github.com/gin-gonic/gin"
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

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "email"})
	})

	r.POST("/send/confirmation", emailHandler.SendConfirmationEmail)
	r.POST("/send/weather-update", emailHandler.SendWeatherUpdateEmail)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Failed to start server: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		panic("Server forced to shutdown: " + err.Error())
	}
}
