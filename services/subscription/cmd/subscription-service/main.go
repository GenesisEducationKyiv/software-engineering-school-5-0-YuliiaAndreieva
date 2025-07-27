package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandler "subscription-service/internal/adapter/http"
	"subscription-service/internal/adapter/logger"
	"subscription-service/internal/adapter/database"
	"subscription-service/internal/config"
	"subscription-service/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	loggerInstance := logger.NewLogrusLogger()

	db, err := gorm.Open(sqlite.Open("subscriptions.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&database.Subscription{})

	repo := database.NewSubscriptionRepo(db)

	subscribeUseCase := usecase.NewSubscribeUseCase(repo, nil, nil, loggerInstance)
	confirmUseCase := usecase.NewConfirmSubscriptionUseCase(repo, nil, loggerInstance)
	unsubscribeUseCase := usecase.NewUnsubscribeUseCase(repo, nil, loggerInstance)

	subscriptionHandler := httphandler.NewSubscriptionHandler(subscribeUseCase, confirmUseCase, unsubscribeUseCase, loggerInstance)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "subscription"})
	})

	r.POST("/subscribe", subscriptionHandler.Subscribe)
	r.GET("/confirm/:token", subscriptionHandler.Confirm)
	r.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)

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