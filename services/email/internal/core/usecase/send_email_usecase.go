package usecase

import (
	"context"
	"email/internal/core/domain"
	"email/internal/core/ports/in"
	"email/internal/core/ports/out"
	"fmt"
)

type SendEmailUseCase struct {
	emailSender     out.EmailSender
	templateBuilder out.EmailTemplateBuilder
	logger          out.Logger
}

func NewSendEmailUseCase(
	emailSender out.EmailSender,
	templateBuilder out.EmailTemplateBuilder,
	logger out.Logger,
) in.SendEmailUseCase {
	return &SendEmailUseCase{
		emailSender:     emailSender,
		templateBuilder: templateBuilder,
		logger:          logger,
	}
}

func (uc *SendEmailUseCase) SendEmail(ctx context.Context, req domain.SendEmailRequest) (*domain.EmailDeliveryResult, error) {
	uc.logger.Infof("Starting email send to %s, type: %s", req.To, req.Type)

	template, err := uc.templateBuilder.BuildTemplate(ctx, req.Type, req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to build template: %w", err)
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

	uc.logger.Infof("Email sent successfully to %s, type: %s", req.To, req.Type)
	return result, nil
}
