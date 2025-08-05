package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	pb "proto/weather"
	"syscall"
	"weather/internal/adapter/cache/core/metrics"
	"weather/internal/adapter/cache/core/redis"
	grpchandler "weather/internal/adapter/grpc"
	"weather/internal/adapter/weather/openweathermap"
	filelogger "weather/internal/utils/logger"

	weathercache "weather/internal/adapter/cache/weather"
	httphandler "weather/internal/adapter/http"

	sharedlogger "shared/logger"
	"weather/internal/adapter/weather"
	"weather/internal/adapter/weather/weatherapi"
	"weather/internal/config"
	"weather/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fileLogger, err := filelogger.NewFileLogger("logs", "provider_responses.log")
	if err != nil {
		log.Fatalf("Unable to initialize file logger: %v", err)
	}
	defer func() {
		if closeErr := fileLogger.Close(); closeErr != nil {
			log.Printf("Error closing file logger: %v", closeErr)
		}
	}()

	baseLogger := sharedlogger.NewZapLoggerWithSampling(cfg.LogInitial, cfg.LogThereafter, cfg.LogTick)

	validateConfig(cfg, baseLogger)

	httpClient := &http.Client{Timeout: cfg.HTTPClientTimeout}

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

	chainProvider := weather.NewChainWeatherProvider(weatherAPIProvider, openWeatherMapProvider)

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
	weatherCache := weathercache.NewCache(cacheWithMetrics)

	cachedProvider := weather.NewCachedWeatherProvider(weatherCache, chainProvider)

	getWeatherUseCase := usecase.NewGetWeatherUseCase(cachedProvider, baseLogger)

	weatherHandler := httphandler.NewWeatherHandler(
		getWeatherUseCase,
		baseLogger,
	)

	grpcHandler := grpchandler.NewWeatherHandler(getWeatherUseCase)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "weather"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/weather", weatherHandler.GetWeather)

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterWeatherServiceServer(grpcSrv, grpcHandler)

	go func() {
		baseLogger.Infof("Starting HTTP server on port %d", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			baseLogger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			baseLogger.Fatalf("Failed to listen for gRPC: %v", err)
		}

		baseLogger.Infof("Starting gRPC server on port %d", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			baseLogger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		baseLogger.Fatalf("HTTP server forced to shutdown: %v", err)
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
