package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/repository"
)

type subscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) repository.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) CreateSubscription(ctx context.Context, sub domain.Subscription) error {
	log.Printf("Creating subscription for city")
	query := `INSERT INTO subscriptions (email, city_id, frequency, token, is_confirmed) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, sub.Email, sub.CityID, sub.Frequency, sub.Token, sub.IsConfirmed)
	if err != nil {
		log.Printf("Unable to create subscription: %v", err)
		return err
	}
	log.Printf("Successfully created subscription")
	return nil
}

func (r *subscriptionRepository) GetSubscriptionByToken(ctx context.Context, token string) (domain.Subscription, error) {
	log.Printf("Looking up subscription")
	var sub domain.Subscription
	query := `SELECT token FROM subscriptions WHERE token = $1`
	err := r.db.QueryRowContext(ctx, query, token).Scan(&sub.Token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No subscription found")
			return domain.Subscription{}, errors.New("subscription not found")
		}
		log.Printf("Error getting subscription: %v", err)
		return domain.Subscription{}, err
	}
	log.Printf("Found subscription")
	return sub, nil
}

func (r *subscriptionRepository) UpdateSubscription(ctx context.Context, sub domain.Subscription) error {
	log.Printf("Updating subscription")
	query := `UPDATE subscriptions SET is_confirmed = $1 WHERE token = $2`
	result, err := r.db.ExecContext(ctx, query, sub.IsConfirmed, sub.Token)
	if err != nil {
		log.Printf("Unable to update subscription: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No subscription found to update")
		return errors.New("subscription not found")
	}

	log.Printf("Successfully updated subscription")
	return nil
}

func (r *subscriptionRepository) DeleteSubscription(ctx context.Context, token string) error {
	log.Printf("Deleting subscription")
	query := `DELETE FROM subscriptions WHERE token = $1`
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		log.Printf("Unable to delete subscription: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No subscription found to delete")
		return errors.New("subscription not found")
	}

	log.Printf("Successfully deleted subscription")
	return nil
}

func (r *subscriptionRepository) GetSubscriptionsByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error) {
	query := `
        SELECT s.id, s.email, s.city_id, c.name as city_name,
               s.frequency, s.token, s.is_confirmed
        FROM subscriptions s
        JOIN cities c ON s.city_id = c.id
        WHERE s.frequency = $1 AND s.is_confirmed = true
    `
	rows, err := r.db.QueryContext(ctx, query, frequency)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var subscriptions []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		var city domain.City
		err := rows.Scan(
			&sub.ID,
			&sub.Email,
			&sub.CityID,
			&city.Name,
			&sub.Frequency,
			&sub.Token,
			&sub.IsConfirmed,
		)
		if err != nil {
			return nil, err
		}
		sub.City = &city
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) IsTokenExists(ctx context.Context, token string) (bool, error) {
	log.Printf("Checking if token exists: %s", token)
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE token = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		log.Printf("Unable to check token existence: %v", err)
		return false, err
	}
	log.Printf("Token existence check result: %v", exists)
	return exists, nil
}

func (r *subscriptionRepository) IsSubscriptionExists(ctx context.Context, opts repository.IsSubscriptionExistsOptions) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1 FROM subscriptions 
            WHERE email = $1 AND city_id = $2 AND frequency = $3
        )
    `
	err := r.db.QueryRowContext(ctx, query, opts.Email, opts.CityID, opts.Frequency).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
