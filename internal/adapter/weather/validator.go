package weather

import (
	"errors"
	"log"
	"strconv"
	"weather-api/internal/core/domain"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateResponse(data WeatherResponse) error {
	if data.Error.Code != 0 {
		if data.Error.Code == 1006 {
			return domain.ErrCityNotFound
		}
		log.Printf("weatherapi error: %s (code %d)", data.Error.Message, data.Error.Code)
		msg := "weatherapi error: " + strconv.Itoa(data.Error.Code) + " " + data.Error.Message
		return errors.New(msg)
	}
	return nil
}
