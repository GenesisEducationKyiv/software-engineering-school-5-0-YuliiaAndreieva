package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscription-service/internal/adapter/database"
	httphandler "subscription-service/internal/adapter/http"
	"subscription-service/internal/adapter/logger"
	"subscription-service/internal/config"
	"subscription-service/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	loggerInstance := logger.NewLogrusLogger()

	var db *gorm.DB
	var err error

	// Retry connection to database
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
		if err == nil {
			break
		}
		loggerInstance.Warnf("Failed to connect to database, retrying in 2 seconds... (attempt %d/30)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database after 30 attempts: %v", err))
	}

	db.AutoMigrate(&database.Subscription{})

	repo := database.NewSubscriptionRepo(db)

	// Create HTTP clients
	httpClient := &http.Client{Timeout: 10 * time.Second}
	emailClient := httphandler.NewEmailClient(cfg.Email.ServiceURL, httpClient, loggerInstance)
	tokenClient := httphandler.NewTokenClient("http://token-service:8083", httpClient, loggerInstance)

	subscribeUseCase := usecase.NewSubscribeUseCase(repo, tokenClient, emailClient, loggerInstance)
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

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "subscription"})
	})

	r.POST("/subscribe", subscriptionHandler.Subscribe)
	r.GET("/confirm/:token", subscriptionHandler.Confirm)
	r.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
	r.POST("/subscriptions/list", subscriptionHandler.ListByFrequency)

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
