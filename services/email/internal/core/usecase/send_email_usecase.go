package usecase

import (
	"context"

	"email/internal/core/domain"
	"email/internal/core/ports/in"
	"email/internal/core/ports/out"
)

type SendEmailUseCase struct {
	emailSender     out.EmailSender
	templateBuilder out.EmailTemplateBuilder
	logger          out.Logger
	baseURL         string
}

func NewSendEmailUseCase(
	emailSender out.EmailSender,
	templateBuilder out.EmailTemplateBuilder,
	logger out.Logger,
	baseURL string,
) in.SendEmailUseCase {
	return &SendEmailUseCase{
		emailSender:     emailSender,
		templateBuilder: templateBuilder,
		logger:          logger,
		baseURL:         baseURL,
	}
}

func (uc *SendEmailUseCase) SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error) {
	uc.logger.Infof("Starting confirmation email send to %s", req.To)

	template, err := uc.templateBuilder.BuildConfirmationEmail(ctx, req.To, req.City, req.ConfirmationLink)
	if err != nil {
		return nil, err
	}

	emailReq := domain.EmailRequest{
		To:      req.To,
		Subject: req.Subject,
		Body:    template,
	}

	result, err := uc.emailSender.SendEmail(ctx, emailReq)
	if err != nil {
		return result, err
	}

	uc.logger.Infof("Confirmation email sent successfully to %s", req.To)
	return result, nil
}

func (uc *SendEmailUseCase) SendWeatherUpdateEmail(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error) {
	uc.logger.Infof("Starting weather update email send to %s", req.To)

	template, err := uc.templateBuilder.BuildWeatherUpdateEmail(ctx, req.To, req.City, req.Description, req.Humidity, req.WindSpeed, req.Temperature, req.UnsubscribeToken)
	if err != nil {
		return nil, err
	}

	emailReq := domain.EmailRequest{
		To:      req.To,
		Subject: req.Subject,
		Body:    template,
	}

	result, err := uc.emailSender.SendEmail(ctx, emailReq)
	if err != nil {
		return result, err
	}

	uc.logger.Infof("Weather update email sent successfully to %s", req.To)
	return result, nil
}
