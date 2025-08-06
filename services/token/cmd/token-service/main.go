package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	sharedlogger "shared/logger"
	httphandler "token/internal/adapter/http"
	"token/internal/config"
	"token/internal/core/ports/in"
	"token/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupUseCases(cfg *config.Config, logger sharedlogger.Logger) (in.GenerateTokenUseCase, in.ValidateTokenUseCase) {
	generateTokenUseCase := usecase.NewGenerateTokenUseCase(logger, cfg.JWT.Secret)
	validateTokenUseCase := usecase.NewValidateTokenUseCase(logger, cfg.JWT.Secret)
	return generateTokenUseCase, validateTokenUseCase
}

func setupHandler(generateTokenUseCase in.GenerateTokenUseCase, validateTokenUseCase in.ValidateTokenUseCase, logger sharedlogger.Logger) *httphandler.TokenHandler {
	return httphandler.NewTokenHandler(generateTokenUseCase, validateTokenUseCase, logger)
}

func setupHTTPServer(cfg *config.Config, tokenHandler *httphandler.TokenHandler) *http.Server {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "token"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/generate", tokenHandler.GenerateToken)
	r.POST("/validate", tokenHandler.ValidateToken)

	return &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}
}

func startServer(srv *http.Server, logger sharedlogger.Logger) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()
}

func gracefulShutdown(srv *http.Server, cfg *config.Config, logger sharedlogger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := sharedlogger.NewZapLoggerWithSampling(cfg.Logging.Initial, cfg.Logging.Thereafter, cfg.Logging.Tick)

	generateTokenUseCase, validateTokenUseCase := setupUseCases(cfg, loggerInstance)
	tokenHandler := setupHandler(generateTokenUseCase, validateTokenUseCase, loggerInstance)

	srv := setupHTTPServer(cfg, tokenHandler)

	startServer(srv, loggerInstance)

	gracefulShutdown(srv, cfg, loggerInstance)
}
