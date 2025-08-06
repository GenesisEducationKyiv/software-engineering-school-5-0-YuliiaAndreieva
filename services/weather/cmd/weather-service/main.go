package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	pb "proto/weather"
	sharedlogger "shared/logger"

	"weather/internal/adapter/cache/core/metrics"
	"weather/internal/adapter/cache/core/redis"
	weathercache "weather/internal/adapter/cache/weather"
	grpchandler "weather/internal/adapter/grpc"
	httphandler "weather/internal/adapter/http"
	"weather/internal/adapter/weather"
	"weather/internal/adapter/weather/openweathermap"
	"weather/internal/adapter/weather/weatherapi"
	"weather/internal/config"
	"weather/internal/core/ports/in"
	"weather/internal/core/ports/out"
	"weather/internal/core/usecase"
	filelogger "weather/internal/utils/logger"
)

func setupFileLogger() out.ProviderLogger {
	fileLogger, err := filelogger.NewFileLogger("logs", "provider_responses.log")
	if err != nil {
		log.Fatalf("Unable to initialize file logger: %v", err)
	}
	return fileLogger
}

func setupWeatherProviders(cfg *config.Config, httpClient *http.Client, fileLogger out.ProviderLogger) weather.Provider {
	weatherAPIProvider := weatherapi.NewClient(weatherapi.ClientOptions{
		APIKey:     cfg.WeatherAPIKey,
		BaseURL:    cfg.WeatherAPIBaseURL,
		HTTPClient: httpClient,
		Logger:     fileLogger,
	})

	openWeatherMapProvider := openweathermap.NewClient(openweathermap.ClientOptions{
		APIKey:     cfg.OpenWeatherMapAPIKey,
		BaseURL:    cfg.OpenWeatherMapBaseURL,
		HTTPClient: httpClient,
		Logger:     fileLogger,
	})

	return weather.NewChainWeatherProvider(weatherAPIProvider, openWeatherMapProvider)
}

func setupCache(cfg *config.Config) *weathercache.Cache {
	redisCache := redis.NewCache(redis.CacheOptions{
		Address:      cfg.RedisAddress,
		TTL:          cfg.RedisTTL,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})

	promRegistry := prometheus.NewRegistry()
	cacheMetrics := weathercache.NewCacheMetrics(promRegistry)
	cacheWithMetrics := metrics.NewCacheWithMetrics(redisCache, cacheMetrics)
	return weathercache.NewCache(cacheWithMetrics)
}

func setupUseCases(chainProvider weather.Provider, weatherCache *weathercache.Cache, logger sharedlogger.Logger) in.GetWeatherUseCase {
	cachedProvider := weather.NewCachedWeatherProvider(weatherCache, chainProvider)
	return usecase.NewGetWeatherUseCase(cachedProvider, logger)
}

func setupHandlers(getWeatherUseCase in.GetWeatherUseCase, logger sharedlogger.Logger) (*httphandler.WeatherHandler, *grpchandler.WeatherHandler) {
	weatherHandler := httphandler.NewWeatherHandler(getWeatherUseCase, logger)
	grpcHandler := grpchandler.NewWeatherHandler(getWeatherUseCase)
	return weatherHandler, grpcHandler
}

func setupHTTPServer(cfg *config.Config, weatherHandler *httphandler.WeatherHandler) *http.Server {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "weather"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/weather", weatherHandler.GetWeather)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}
}

func setupGRPCServer(grpcHandler *grpchandler.WeatherHandler) *grpc.Server {
	grpcSrv := grpc.NewServer()
	pb.RegisterWeatherServiceServer(grpcSrv, grpcHandler)
	return grpcSrv
}

func startServers(httpSrv *http.Server, grpcSrv *grpc.Server, cfg *config.Config, logger sharedlogger.Logger) {
	go func() {
		logger.Infof("Starting HTTP server on port %d", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			logger.Fatalf("Failed to listen for gRPC: %v", err)
		}

		logger.Infof("Starting gRPC server on port %d", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
}

func gracefulShutdown(httpSrv *http.Server, grpcSrv *grpc.Server, cfg *config.Config, logger sharedlogger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Fatalf("HTTP server forced to shutdown: %v", err)
	}
}

func validateConfig(cfg *config.Config, logger sharedlogger.Logger) {
	if cfg.BaseURL == "" {
		logger.Fatalf("BASE_URL environment variable is required")
	}
	if cfg.Port == 0 {
		logger.Fatalf("PORT environment variable is required")
	}
	if cfg.GRPCPort == 0 {
		logger.Fatalf("GRPC_PORT environment variable is required")
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fileLogger := setupFileLogger()
	defer func() {
		if closeErr := fileLogger.Close(); closeErr != nil {
			log.Printf("Error closing file logger: %v", closeErr)
		}
	}()

	baseLogger := sharedlogger.NewZapLoggerWithSampling(cfg.LogInitial, cfg.LogThereafter, cfg.LogTick)

	validateConfig(cfg, baseLogger)

	httpClient := &http.Client{Timeout: cfg.HTTPClientTimeout}

	chainProvider := setupWeatherProviders(cfg, httpClient, fileLogger)
	weatherCache := setupCache(cfg)
	getWeatherUseCase := setupUseCases(chainProvider, weatherCache, baseLogger)

	weatherHandler, grpcHandler := setupHandlers(getWeatherUseCase, baseLogger)

	httpSrv := setupHTTPServer(cfg, weatherHandler)
	grpcSrv := setupGRPCServer(grpcHandler)

	startServers(httpSrv, grpcSrv, cfg, baseLogger)

	gracefulShutdown(httpSrv, grpcSrv, cfg, baseLogger)
}
