package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports/out"
)

type CityService struct {
	cityRepo        out.CityRepository
	weatherProvider out.WeatherProvider
}

func NewCityService(
	cityRepo out.CityRepository,
	weatherProvider out.WeatherProvider,
) *CityService {
	return &CityService{
		cityRepo:        cityRepo,
		weatherProvider: weatherProvider,
	}
}

func (s *CityService) EnsureCityExists(ctx context.Context, cityName string) (domain.City, error) {
	city, err := s.cityRepo.GetByName(ctx, cityName)
	if err == nil {
		log.Printf("City %s already exists in database", cityName)
		return city, nil
	}

	if !errors.Is(err, domain.ErrCityNotFound) {
		return domain.City{}, fmt.Errorf("unable to get city %s: %w", cityName, err)
	}

	if err := s.weatherProvider.CheckCityExists(ctx, cityName); err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			log.Printf("City %s not found in weather service", cityName)
			return domain.City{}, domain.ErrCityNotFound
		}
		return domain.City{}, fmt.Errorf("unable to check city existence for %s: %w", cityName, err)
	}

	city, err = s.cityRepo.Create(ctx, domain.City{Name: cityName})
	if err != nil {
		return domain.City{}, fmt.Errorf("unable to create city %s: %w", cityName, err)
	}

	log.Printf("Successfully created city: %s with ID: %d", cityName, city.ID)
	return city, nil
}
