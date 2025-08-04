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
	"time"
	"weather/internal/adapter/cache/core/metrics"
	"weather/internal/adapter/cache/core/redis"
	grpchandler "weather/internal/adapter/grpc"
	"weather/internal/adapter/weather/openweathermap"
	"weather/internal/utils/logger"

	weathercache "weather/internal/adapter/cache/weather"
	httphandler "weather/internal/adapter/http"

	"weather/internal/adapter/weather"
	"weather/internal/adapter/weather/weatherapi"
	"weather/internal/config"
	"weather/internal/core/usecase"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	fileLogger, err := logger.NewFileLogger("logs", "provider_responses.log")
	if err != nil {
		log.Fatalf("Unable to initialize file logger: %v", err)
	}
	defer func() {
		if closeErr := fileLogger.Close(); closeErr != nil {
			log.Printf("Error closing file logger: %v", closeErr)
		}
	}()

	appLogger := logger.NewLogger()

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

	getWeatherUseCase := usecase.NewGetWeatherUseCase(cachedProvider, appLogger)

	weatherHandler := httphandler.NewWeatherHandler(
		getWeatherUseCase,
		appLogger,
	)

	grpcHandler := grpchandler.NewWeatherHandler(getWeatherUseCase)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "weather"})
	})

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
		appLogger.Infof("Starting HTTP server on port %d", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Failed to start HTTP server: " + err.Error())
		}
	}()

	go func() {
		grpcPort := cfg.GRPCPort
		if grpcPort == 0 {
			grpcPort = 9092
		}
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			panic(fmt.Sprintf("Failed to listen for gRPC: %v", err))
		}

		appLogger.Infof("Starting gRPC server on port %d", grpcPort)
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
