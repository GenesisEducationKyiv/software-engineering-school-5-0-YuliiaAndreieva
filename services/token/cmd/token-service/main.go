package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	sharedlogger "shared/logger"
	httphandler "token/internal/adapter/http"
	"token/internal/config"
	"token/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.LoadConfig()

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	validateConfig(cfg, loggerInstance)

	generateTokenUseCase := usecase.NewGenerateTokenUseCase(loggerInstance, cfg.JWT.Secret)
	validateTokenUseCase := usecase.NewValidateTokenUseCase(loggerInstance, cfg.JWT.Secret)

	tokenHandler := httphandler.NewTokenHandler(generateTokenUseCase, validateTokenUseCase, loggerInstance)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "token"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/generate", tokenHandler.GenerateToken)
	r.POST("/validate", tokenHandler.ValidateToken)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInstance.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		loggerInstance.Fatalf("Server forced to shutdown: %v", err)
	}
}

func validateConfig(cfg *config.Config, logger sharedlogger.Logger) {
	if cfg.JWT.Secret == "" {
		logger.Fatalf("JWT_SECRET environment variable is required")
	}
	if cfg.Server.Port == "" {
		logger.Fatalf("SERVER_PORT environment variable is required")
	}
}
