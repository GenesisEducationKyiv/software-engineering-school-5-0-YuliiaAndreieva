package grpc

import (
	"context"
	"fmt"
	pb "proto/weather"
	sharedlogger "shared/logger"
	"weather-broadcast/internal/core/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WeatherClient struct {
	client pb.WeatherServiceClient
	logger sharedlogger.Logger
}

func NewWeatherClient(address string, logger sharedlogger.Logger) (*WeatherClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to weather service: %w", err)
	}

	client := pb.NewWeatherServiceClient(conn)
	return &WeatherClient{
		client: client,
		logger: logger,
	}, nil
}

func (c *WeatherClient) GetWeatherByCity(ctx context.Context, city string) (*domain.Weather, error) {
	c.logger.Debugf("Getting weather for city: %s", city)

	req := &pb.WeatherRequest{
		City: city,
	}

	resp, err := c.client.GetWeatherByCity(ctx, req)
	if err != nil {
		c.logger.Errorf("Failed to get weather: %v", err)
		return nil, fmt.Errorf("failed to get weather: %w", err)
	}

	weather := &domain.Weather{
		Temperature: resp.Weather.Temperature,
		Humidity:    int(resp.Weather.Humidity),
		Description: resp.Weather.Description,
		WindSpeed:   resp.Weather.WindSpeed,
	}

	c.logger.Infof("Successfully retrieved weather for city: %s", city)
	return weather, nil
}
