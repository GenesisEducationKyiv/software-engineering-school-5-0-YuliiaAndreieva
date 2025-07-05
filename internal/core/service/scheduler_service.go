package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
)

type WeatherUpdateService interface {
	PrepareUpdates(ctx context.Context, frequency domain.Frequency) ([]domain.WeatherUpdate, error)
}

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
		msg := fmt.Sprintf("unable to prepare updates for frequency %s: %v", frequency, err)
		log.Print(msg)
		return errors.New(msg)
	}

	if err := s.emailService.SendUpdates(updates); err != nil {
		msg := fmt.Sprintf("unable to send updates for frequency %s: %v", frequency, err)
		log.Print(msg)
		return errors.New(msg)
	}

	return nil
}
