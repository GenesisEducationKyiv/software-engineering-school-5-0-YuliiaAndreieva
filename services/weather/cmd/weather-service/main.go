package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather-service/internal/adapter/cache/core/metrics"
	"weather-service/internal/adapter/cache/core/redis"
	"weather-service/internal/adapter/weather/openweathermap"
	"weather-service/internal/utils/logger"

	weathercache "weather-service/internal/adapter/cache/weather"
	httphandler "weather-service/internal/adapter/http"

	"weather-service/internal/adapter/weather"
	"weather-service/internal/adapter/weather/weatherapi"
	"weather-service/internal/config"
	"weather-service/internal/core/usecase"

	"github.com/gin-gonic/gin"
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

	getWeatherUseCase := usecase.NewGetWeatherUseCase(cachedProvider)

	weatherHandler := httphandler.NewWeatherHandler(
		getWeatherUseCase,
	)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "weather"})
	})

	r.POST("/weather", weatherHandler.GetWeather)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
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
