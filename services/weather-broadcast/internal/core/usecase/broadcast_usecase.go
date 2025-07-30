package usecase

import (
	"context"
	"sync"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/internal/core/ports/in"
	"weather-broadcast/internal/core/ports/out"
)

const (
	PAGE_SIZE     = 100
	WORKER_AMOUNT = 10
)

type BroadcastUseCase struct {
	subscriptionClient out.SubscriptionClient
	weatherClient      out.WeatherClient
	weatherMailer      out.WeatherMailer
	logger             out.Logger
}

func NewBroadcastUseCase(
	subscriptionClient out.SubscriptionClient,
	weatherClient out.WeatherClient,
	weatherMailer out.WeatherMailer,
	logger out.Logger,
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
	sem := make(chan struct{}, WORKER_AMOUNT)
	wg := &sync.WaitGroup{}

	lastID := 0
	for {
		query := domain.ListSubscriptionsQuery{
			Frequency: frequency,
			LastID:    lastID,
			PageSize:  PAGE_SIZE,
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

		for _, subscription := range subscriptions {
			if _, ok := cityWeatherMap[subscription.City]; !ok {
				weather, err := s.weatherClient.GetWeatherByCity(ctx, subscription.City)
				s.logger.Infof("Weather for %s: %v", subscription.City, weather)

				if err != nil {
					cityWeatherMap[subscription.City] = nil
				} else {
					cityWeatherMap[subscription.City] = weather
				}
			}

			sem <- struct{}{}
			wg.Add(1)

			go func(sub domain.Subscription, weather *domain.Weather) {
				defer func() { <-sem }()
				defer wg.Done()

				if weather != nil {
					info := &domain.WeatherMailSuccessInfo{
						Email:   sub.Email,
						City:    sub.City,
						Weather: *weather,
					}

					if err := s.weatherMailer.SendWeather(ctx, info); err != nil {
						s.logger.Errorf("Failed to send weather email to %s: %v", sub.Email, err)
					}
				} else {
					info := &domain.WeatherMailErrorInfo{
						Email: sub.Email,
						City:  sub.City,
					}

					if err := s.weatherMailer.SendError(ctx, info); err != nil {
						s.logger.Errorf("Failed to send error email to %s: %v", sub.Email, err)
					}
				}
			}(subscription, cityWeatherMap[subscription.City])
		}
	}
	wg.Wait()

	s.logger.Infof("Broadcast completed for frequency: %s", frequency)
	return nil
}
