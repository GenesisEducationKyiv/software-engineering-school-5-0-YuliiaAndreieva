package strategies

import (
	"email/internal/core/domain"
	"email/internal/core/ports/out"
)

type StrategyFactory struct {
	logger                 out.Logger
	subscriptionServiceURL string
}

func NewStrategyFactory(logger out.Logger, subscriptionServiceURL string) *StrategyFactory {
	return &StrategyFactory{
		logger:                 logger,
		subscriptionServiceURL: subscriptionServiceURL,
	}
}

func (f *StrategyFactory) CreateStrategies() map[domain.EmailType]out.EmailTemplateStrategy {
	strategies := make(map[domain.EmailType]out.EmailTemplateStrategy)

	strategies[domain.EmailTypeConfirmation] = NewConfirmationEmailStrategy(f.logger)
	strategies[domain.EmailTypeWeatherUpdate] = NewWeatherUpdateEmailStrategy(f.logger, f.subscriptionServiceURL)

	return strategies
}
