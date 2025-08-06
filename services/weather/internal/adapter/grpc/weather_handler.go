package grpc

import (
	"context"
	pb "proto/weather"
	"weather/internal/core/domain"
	"weather/internal/core/ports/in"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WeatherHandler struct {
	pb.UnimplementedWeatherServiceServer
	getWeatherUseCase in.GetWeatherUseCase
}

func NewWeatherHandler(getWeatherUseCase in.GetWeatherUseCase) *WeatherHandler {
	return &WeatherHandler{
		getWeatherUseCase: getWeatherUseCase,
	}
}

func (h *WeatherHandler) GetWeatherByCity(ctx context.Context, req *pb.WeatherRequest) (*pb.WeatherResponse, error) {
	weatherReq := domain.WeatherRequest{
		City: req.City,
	}

	weatherResp, err := h.getWeatherUseCase.GetWeather(ctx, weatherReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get weather: %v", err)
	}

	return &pb.WeatherResponse{
		Weather: &pb.Weather{
			Temperature: weatherResp.Weather.Temperature,
			Humidity:    int32(weatherResp.Weather.Humidity),
			Description: weatherResp.Weather.Description,
			WindSpeed:   weatherResp.Weather.WindSpeed,
		},
	}, nil
}
