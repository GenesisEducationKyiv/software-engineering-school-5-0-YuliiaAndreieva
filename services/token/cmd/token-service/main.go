package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandler "token/internal/adapter/http"
	"token/internal/adapter/logger"
	"token/internal/config"
	"token/internal/core/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	loggerInstance := logger.NewLogrusLogger()

	generateTokenUseCase := usecase.NewGenerateTokenUseCase(loggerInstance, cfg.JWT.Secret)
	validateTokenUseCase := usecase.NewValidateTokenUseCase(loggerInstance, cfg.JWT.Secret)

	tokenHandler := httphandler.NewTokenHandler(generateTokenUseCase, validateTokenUseCase, loggerInstance)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "token"})
	})

	r.POST("/generate", tokenHandler.GenerateToken)
	r.POST("/validate", tokenHandler.ValidateToken)

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
