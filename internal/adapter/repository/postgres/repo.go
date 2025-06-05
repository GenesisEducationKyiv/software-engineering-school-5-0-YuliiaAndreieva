package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/port"
)

type SubscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) port.SubscriptionRepository {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) CreateSubscription(ctx context.Context, sub domain.Subscription) error {
	log.Printf("Creating subscription for city: %s", sub.City)
	query := `INSERT INTO subscriptions (email, city, frequency, token, is_confirmed) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, sub.Email, sub.City, sub.Frequency, sub.Token, sub.IsConfirmed)
	if err != nil {
		log.Printf("Failed to create subscription: %v", err)
		return err
	}
	log.Printf("Successfully created subscription")
	return nil
}

func (r *SubscriptionRepo) GetSubscriptionByToken(ctx context.Context, token string) (domain.Subscription, error) {
	log.Printf("Looking up subscription")
	var sub domain.Subscription
	query := `SELECT id, email, city, frequency, token, is_confirmed FROM subscriptions WHERE token = $1`
	err := r.db.QueryRowContext(ctx, query, token).Scan(&sub.ID, &sub.Email, &sub.City, &sub.Frequency, &sub.Token, &sub.IsConfirmed)
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

func (r *SubscriptionRepo) UpdateSubscription(ctx context.Context, sub domain.Subscription) error {
	log.Printf("Updating subscription")
	query := `UPDATE subscriptions SET is_confirmed = $1 WHERE token = $2`
	result, err := r.db.ExecContext(ctx, query, sub.IsConfirmed, sub.Token)
	if err != nil {
		log.Printf("Failed to update subscription: %v", err)
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

func (r *SubscriptionRepo) DeleteSubscription(ctx context.Context, token string) error {
	log.Printf("Deleting subscription")
	query := `DELETE FROM subscriptions WHERE token = $1`
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		log.Printf("Failed to delete subscription: %v", err)
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

func (r *SubscriptionRepo) GetSubscriptionsByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error) {
	log.Printf("Getting subscriptions for frequency: %s", frequency)
	query := `SELECT id, email, city, frequency, token, is_confirmed FROM subscriptions WHERE frequency = $1 AND is_confirmed = true`
	rows, err := r.db.QueryContext(ctx, query, frequency)
	if err != nil {
		log.Printf("Failed to query subscriptions: %v", err)
		return nil, err
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.Email, &sub.City, &sub.Frequency, &sub.Token, &sub.IsConfirmed); err != nil {
			log.Printf("Error scanning subscription row: %v", err)
			return nil, err
		}
		subs = append(subs, sub)
	}

	log.Printf("Found %d subscriptions for frequency: %s", len(subs), frequency)
	return subs, nil
}

func (r *SubscriptionRepo) IsEmailSubscribed(ctx context.Context, email string) (bool, error) {
	log.Printf("Checking if email is already subscribed: %s", email)
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		log.Printf("Failed to check email subscription: %v", err)
		return false, err
	}
	log.Printf("Email subscription check result: %v", exists)
	return exists, nil
}

func (r *SubscriptionRepo) IsTokenExists(ctx context.Context, token string) (bool, error) {
	log.Printf("Checking if token exists: %s", token)
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE token = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		log.Printf("Failed to check token existence: %v", err)
		return false, err
	}
	log.Printf("Token existence check result: %v", exists)
	return exists, nil
}
