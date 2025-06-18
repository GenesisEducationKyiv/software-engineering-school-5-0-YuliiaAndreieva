package service

import (
	"log"
	"weather-api/internal/adapter/email"
	"weather-api/internal/core/domain"
	"weather-api/internal/util/emailutil"
)

type EmailService struct {
	emailSvc email.EmailSender
}

func NewEmailService(emailSvc email.EmailSender) *EmailService {
	return &EmailService{
		emailSvc: emailSvc,
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

func (s *EmailService) SendConfirmationEmail(subscription *domain.Subscription) error {
	subject, htmlBody := emailutil.BuildConfirmationEmail(subscription.City.Name, subscription.Token)

	if err := s.emailSvc.SendEmail(subscription.Email, subject, htmlBody); err != nil {
		log.Printf("Failed to send confirmation email to %s: %v", subscription.Email, err)
		return err
	}

	return nil
}
