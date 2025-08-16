package usecase

import (
	"context"
	sharedlogger "shared/logger"
	"sync"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/internal/core/ports/in"
	"weather-broadcast/internal/core/ports/out"
)

const (
	PageSize     = 100
	WorkerAmount = 10
)

type BroadcastUseCase struct {
	subscriptionClient out.SubscriptionClient
	weatherClient      out.WeatherClient
	weatherMailer      out.WeatherMailer
	logger             sharedlogger.Logger
}

func NewBroadcastUseCase(
	subscriptionClient out.SubscriptionClient,
	weatherClient out.WeatherClient,
	weatherMailer out.WeatherMailer,
	logger sharedlogger.Logger,
) in.BroadcastUseCase {
	return &BroadcastUseCase{
		subscriptionClient: subscriptionClient,
		weatherClient:      weatherClient,
		weatherMailer:      weatherMailer,
		logger:             logger,
	}
}

func (s *BroadcastUseCase) Broadcast(ctx context.Context, frequency domain.Frequency) error {
	cityWeatherMap := make(map[string]*domain.Weather)
	sem := make(chan struct{}, WorkerAmount)
	wg := &sync.WaitGroup{}

	lastID := 0
	for {
		query := domain.ListSubscriptionsQuery{
			Frequency: frequency,
			LastID:    lastID,
			PageSize:  PageSize,
		}

		res, err := s.subscriptionClient.ListByFrequency(ctx, query)
		if err != nil {
			s.logger.Warnf("Failed to fetch subscriptions: %v", err)
			break
		}

		subscriptions := res.Subscriptions
		lastID = res.LastIndex

		if len(subscriptions) == 0 {
			break
		}

		for _, sub := range subscriptions {
			select {
			case <-ctx.Done():
				s.logger.Warnf("Broadcast canceled during subscription processing: %v", ctx.Err())
				return ctx.Err()
			default:
				sem <- struct{}{}
				wg.Add(1)

				go func(sub domain.Subscription, weather *domain.Weather) {
					defer wg.Done()
					defer func() { <-sem }()

					select {
					case <-ctx.Done():
						s.logger.Warnf("Broadcast canceled during email sending: %v", ctx.Err())
						return
					default:
						s.logger.Infof("Processing subscription for user: %s", sub.Email)
						if err := s.weatherMailer.SendWeather(ctx, &domain.WeatherMailSuccessInfo{
							Email:   sub.Email,
							City:    sub.City,
							Weather: *weather,
							Token:   sub.Token,
						}); err != nil {
							s.logger.Errorf("Failed to send weather email to %s: %v", sub.Email, err)
						}
					}
				}(sub, cityWeatherMap[sub.City])
			}
		}
	}
	wg.Wait()

	s.logger.Infof("Broadcast completed for frequency: %s", frequency)
	return nil
}
