package service

import (
	"errors"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"
	"weather-api/internal/util/emailutil"
)

type EmailService interface {
	SendUpdates(updates []domain.WeatherUpdate) error
	SendConfirmationEmail(subscription *domain.Subscription) error
}

type EmailServiceImpl struct {
	emailSvc ports.EmailSender
}

func NewEmailService(emailSvc ports.EmailSender) *EmailServiceImpl {
	return &EmailServiceImpl{
		emailSvc: emailSvc,
	}
}

func (s *EmailServiceImpl) SendUpdates(updates []domain.WeatherUpdate) error {
	for _, update := range updates {
		subject, htmlBody := emailutil.BuildWeatherUpdateEmail(emailutil.WeatherUpdateEmailOptions{
			City:        update.Subscription.City.Name,
			Temperature: update.Weather.Temperature,
			Humidity:    update.Weather.Humidity,
			Description: update.Weather.Description,
			Token:       update.Subscription.Token,
		})

		if err := s.emailSvc.SendEmail(ports.SendEmailOptions{
			To:      update.Subscription.Email,
			Subject: subject,
			Body:    htmlBody,
		}); err != nil {
			msg := fmt.Sprintf("unable to send email to %s: %v", update.Subscription.Email, err)
			log.Print(msg)
			return errors.New(msg)
		}
	}

	return nil
}

func (s *EmailServiceImpl) SendConfirmationEmail(subscription *domain.Subscription) error {
	subject, htmlBody := emailutil.BuildConfirmationEmail(subscription.City.Name, subscription.Token)

	if err := s.emailSvc.SendEmail(ports.SendEmailOptions{
		To:      subscription.Email,
		Subject: subject,
		Body:    htmlBody,
	}); err != nil {
		msg := fmt.Sprintf("unable to send confirmation email to %s: %v", subscription.Email, err)
		log.Print(msg)
		return errors.New(msg)
	}

	return nil
}
