package grpc

import (
	"context"
	"fmt"
	pb "proto/email"
	sharedlogger "shared/logger"
	"weather-broadcast/internal/core/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EmailClient struct {
	client pb.EmailServiceClient
	logger sharedlogger.Logger
}

func NewEmailClient(address string, logger sharedlogger.Logger) (*EmailClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to email service: %w", err)
	}

	client := pb.NewEmailServiceClient(conn)
	return &EmailClient{
		client: client,
		logger: logger,
	}, nil
}

func (c *EmailClient) SendWeather(ctx context.Context, info *domain.WeatherMailSuccessInfo) error {
	c.logger.Debugf("Sending weather update email to: %s for city: %s", info.Email, info.City)

	req := &pb.WeatherUpdateRequest{
		To:               info.Email,
		Subject:          "Weather Update",
		Name:             "User",
		City:             info.City,
		Description:      info.Weather.Description,
		Temperature:      int32(info.Weather.Temperature),
		Humidity:         int32(info.Weather.Humidity),
		WindSpeed:        int32(info.Weather.WindSpeed),
		UnsubscribeToken: info.Token,
	}

	resp, err := c.client.SendWeatherUpdate(ctx, req)
	if err != nil {
		c.logger.Errorf("Failed to send weather update email: %v", err)
		return fmt.Errorf("failed to send weather update email: %w", err)
	}

	if !resp.Success {
		c.logger.Errorf("Email service returned error: %s", resp.Error)
		return fmt.Errorf("email service error: %s", resp.Error)
	}

	c.logger.Infof("Successfully sent weather update email to: %s for city: %s", info.Email, info.City)
	return nil
}
