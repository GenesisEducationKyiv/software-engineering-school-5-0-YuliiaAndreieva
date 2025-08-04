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
	request := domain.WeatherUpdateEmailRequest{
		To:               req.To,
		Subject:          req.Subject,
		Name:             req.Name,
		City:             req.City,
		Description:      req.Description,
		Temperature:      int(req.Temperature),
		Humidity:         int(req.Humidity),
		WindSpeed:        int(req.WindSpeed),
		UnsubscribeToken: req.UnsubscribeToken,
	}

	_, err := h.sendEmailUseCase.SendWeatherUpdateEmail(ctx, request)
	if err != nil {
		return &pb.EmailResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Errorf(codes.Internal, "failed to send weather update email: %v", err)
	}

	return &pb.EmailResponse{
		Success: true,
		Message: "Weather update email sent successfully",
	}, nil
}
