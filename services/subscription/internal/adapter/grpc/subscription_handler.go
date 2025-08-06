package grpc

import (
	"context"
	pb "proto/subscription"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/in"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubscriptionHandler struct {
	pb.UnimplementedSubscriptionServiceServer
	listByFrequencyUseCase in.ListByFrequencyUseCase
}

func NewSubscriptionHandler(
	listByFrequencyUseCase in.ListByFrequencyUseCase,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		listByFrequencyUseCase: listByFrequencyUseCase,
	}
}

func (h *SubscriptionHandler) ListByFrequency(ctx context.Context, req *pb.ListByFrequencyRequest) (*pb.ListByFrequencyResponse, error) {
	query := domain.ListSubscriptionsQuery{
		Frequency: req.Frequency,
		LastID:    int(req.LastId),
		PageSize:  int(req.PageSize),
	}

	result, err := h.listByFrequencyUseCase.ListByFrequency(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list subscriptions: %v", err)
	}

	subscriptions := make([]*pb.Subscription, len(result.Subscriptions))
	for i, sub := range result.Subscriptions {
		subscriptions[i] = &pb.Subscription{
			Id:        int32(sub.ID),
			Email:     sub.Email,
			City:      sub.City,
			Frequency: sub.Frequency,
			Confirmed: sub.IsConfirmed,
			Token:     sub.Token,
		}
	}

	return &pb.ListByFrequencyResponse{
		Subscriptions: subscriptions,
		LastId:        int32(result.LastIndex),
	}, nil
}
