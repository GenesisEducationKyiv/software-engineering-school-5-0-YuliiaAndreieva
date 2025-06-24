package service

import (
	"context"
	"log"
	"weather-api/internal/core/domain"
)

type SchedulerService struct {
	weatherUpdateService WeatherUpdateService
	emailService         EmailService
}

func NewSchedulerService(weatherUpdateService WeatherUpdateService, emailService EmailService) *SchedulerService {
	return &SchedulerService{
		weatherUpdateService: weatherUpdateService,
		emailService:         emailService,
	}
}

func (s *SchedulerService) SendWeatherUpdates(ctx context.Context, frequency domain.Frequency) error {
	updates, err := s.weatherUpdateService.PrepareUpdates(ctx, frequency)
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
