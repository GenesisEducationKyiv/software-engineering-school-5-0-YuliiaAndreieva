package service

import (
	"context"
	"log"
	"weather-api/internal/core/domain"
)

type SchedulerService struct {
	subscriptionService *SubscriptionService
	emailService        *EmailService
}

func NewSchedulerService(subscriptionService *SubscriptionService, emailService *EmailService) *SchedulerService {
	return &SchedulerService{
		subscriptionService: subscriptionService,
		emailService:        emailService,
	}
}

func (s *SchedulerService) SendWeatherUpdates(ctx context.Context, frequency domain.Frequency) error {
	updates, err := s.subscriptionService.PrepareUpdates(ctx, frequency)
	if err != nil {
		log.Printf("Failed to prepare updates for frequency %s: %v", frequency, err)
		return err
	}

	if err := s.emailService.SendUpdates(updates); err != nil {
		log.Printf("Failed to send updates for frequency %s: %v", frequency, err)
		return err
	}

	return nil
}
