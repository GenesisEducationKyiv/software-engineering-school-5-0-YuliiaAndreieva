package service

import (
	"log"
	"weather-api/internal/adapter/email"
	"weather-api/internal/adapter/repository/postgres"
	"weather-api/internal/core/domain"
	"weather-api/internal/util/emailutil"
)

type EmailService struct {
	repo       postgres.SubscriptionRepository
	weatherSvc WeatherService
	emailSvc   email.EmailSender
}

func NewEmailService(repo postgres.SubscriptionRepository, weatherSvc WeatherService, emailSvc email.EmailSender) *EmailService {
	return &EmailService{
		repo:       repo,
		weatherSvc: weatherSvc,
		emailSvc:   emailSvc,
	}
}

func (s *EmailService) SendUpdates(updates []domain.WeatherUpdate) error {
	for _, update := range updates {
		subject, htmlBody := emailutil.BuildWeatherUpdateEmail(
			update.Subscription.City.Name,
			update.Weather.Temperature,
			update.Weather.Humidity,
			update.Weather.Description,
			update.Subscription.Token,
		)

		if err := s.emailSvc.SendEmail(update.Subscription.Email, subject, htmlBody); err != nil {
			log.Printf("Failed to send email to %s: %v", update.Subscription.Email, err)
			continue
		}
	}

	return nil
}
