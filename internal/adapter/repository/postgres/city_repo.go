package postgres

import (
	"context"
	"database/sql"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/repository"
)

type cityRepo struct {
	db *sql.DB
}

func NewCityRepository(db *sql.DB) repository.CityRepository {
	return &cityRepo{db: db}
}

func (r *cityRepo) Create(ctx context.Context, city domain.City) (domain.City, error) {
	log.Printf("Attempting to create city: %s", city.Name)

	query := `INSERT INTO cities (name) VALUES ($1) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, city.Name).Scan(&city.ID)
	if err != nil {
		log.Printf("Failed to create city %s: %v", city.Name, err)
		return domain.City{}, err
	}

	log.Printf("Successfully created city: %s with ID: %d", city.Name, city.ID)
	return city, nil
}

func (r *cityRepo) GetByName(ctx context.Context, name string) (domain.City, error) {
	log.Printf("Attempting to get city by name: %s", name)

	var city domain.City
	query := `SELECT id, name FROM cities WHERE name = $1`
	err := r.db.QueryRowContext(ctx, query, name).Scan(&city.ID, &city.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("City not found: %s", name)
			return domain.City{}, domain.ErrCityNotFound
		}
		log.Printf("Failed to get city %s: %v", name, err)
		return domain.City{}, err
	}

	log.Printf("Successfully found city: %s with ID: %d", city.Name, city.ID)
	return city, nil
}
