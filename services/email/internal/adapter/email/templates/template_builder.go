package templates

import (
	"context"
	"email/internal/core/domain"
	"email/internal/core/ports/out"
	"fmt"
)

type TemplateBuilder struct {
	strategies map[domain.EmailType]out.EmailTemplateStrategy
	logger     out.Logger
}

func NewTemplateBuilder(
	logger out.Logger,
	strategies map[domain.EmailType]out.EmailTemplateStrategy,
) out.EmailTemplateBuilder {
	return &TemplateBuilder{
		strategies: strategies,
		logger:     logger,
	}
}

func (tb *TemplateBuilder) BuildEmailTemplate(ctx context.Context, req domain.SendEmailRequest) (string, error) {
	strategy, exists := tb.strategies[req.Type]
	if !exists {
		return "", fmt.Errorf("unsupported email type: %s", req.Type)
	}

	template, err := strategy.BuildTemplate(ctx, req.Data)
	if err != nil {
		return "", fmt.Errorf("failed to build template for email type %s: %w", req.Type, err)
	}

	return template, nil
}
