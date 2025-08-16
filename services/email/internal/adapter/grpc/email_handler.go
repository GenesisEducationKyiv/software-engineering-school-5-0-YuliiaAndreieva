package grpc

import (
	"context"
	"email/internal/core/domain"
	"email/internal/core/ports/in"
	pb "proto/email"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EmailHandler struct {
	pb.UnimplementedEmailServiceServer
	sendEmailUseCase in.SendEmailUseCase
}

func NewEmailHandler(sendEmailUseCase in.SendEmailUseCase) *EmailHandler {
	return &EmailHandler{
		sendEmailUseCase: sendEmailUseCase,
	}
}

func (h *EmailHandler) SendWeatherUpdate(ctx context.Context, req *pb.WeatherUpdateRequest) (*pb.EmailResponse, error) {
	request := domain.SendEmailRequest{
		Type: domain.EmailTypeWeatherUpdate,
		Data: map[string]interface{}{
			"city":             req.City,
			"description":      req.Description,
			"temperature":      req.Temperature,
			"humidity":         req.Humidity,
			"windSpeed":        req.WindSpeed,
			"unsubscribeToken": req.UnsubscribeToken,
		},
		To:      req.To,
		Subject: req.Subject,
	}

	result, err := h.sendEmailUseCase.SendEmail(ctx, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send weather update email: %v", err)
	}

	return &pb.EmailResponse{
		To:     result.To,
		SentAt: result.SentAt,
	}, nil
}
