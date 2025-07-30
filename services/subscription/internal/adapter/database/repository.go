package database

import (
	"context"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/out"

	"gorm.io/gorm"
)

type SubscriptionRepo struct {
	db     *gorm.DB
	logger out.Logger
}

func NewSubscriptionRepo(db *gorm.DB, logger out.Logger) out.SubscriptionRepository {
	return &SubscriptionRepo{db: db, logger: logger}
}

func (r *SubscriptionRepo) CreateSubscription(ctx context.Context, s domain.Subscription) error {
	r.logger.Debugf("Creating subscription in database for email: %s, city: %s", s.Email, s.City)

	sub := Subscription{
		Email:     s.Email,
		City:      s.City,
		Token:     s.Token,
		Frequency: Frequency(s.Frequency),
		Confirmed: s.IsConfirmed,
	}

	err := r.db.WithContext(ctx).Create(&sub).Error
	if err != nil {
		r.logger.Errorf("Failed to create subscription in database: %v", err)
		return err
	}

	r.logger.Infof("Successfully created subscription in database for email: %s, city: %s", s.Email, s.City)
	return nil
}

func (r *SubscriptionRepo) GetSubscriptionByToken(ctx context.Context, token string) (*domain.Subscription, error) {
	r.logger.Debugf("Getting subscription by token: %s", token)

	var sub Subscription
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&sub).Error; err != nil {
		r.logger.Errorf("Failed to get subscription by token %s: %v", token, err)
		return nil, err
	}

	r.logger.Debugf("Found subscription by token: %s, email: %s", token, sub.Email)
	return &domain.Subscription{
		ID:          int64(sub.ID),
		Email:       sub.Email,
		City:        sub.City,
		Frequency:   string(sub.Frequency),
		Token:       sub.Token,
		IsConfirmed: sub.Confirmed,
	}, nil
}

func (r *SubscriptionRepo) UpdateSubscription(ctx context.Context, s domain.Subscription) error {
	r.logger.Debugf("Updating subscription for token: %s", s.Token)

	err := r.db.WithContext(ctx).Model(&Subscription{}).
		Where("token = ?", s.Token).
		Updates(map[string]interface{}{
			"confirmed": s.IsConfirmed,
		}).Error

	if err != nil {
		r.logger.Errorf("Failed to update subscription for token %s: %v", s.Token, err)
		return err
	}

	r.logger.Infof("Successfully updated subscription for token: %s", s.Token)
	return nil
}

func (r *SubscriptionRepo) DeleteSubscription(ctx context.Context, token string) error {
	r.logger.Debugf("Deleting subscription for token: %s", token)

	err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&Subscription{}).Error
	if err != nil {
		r.logger.Errorf("Failed to delete subscription for token %s: %v", token, err)
		return err
	}

	r.logger.Infof("Successfully deleted subscription for token: %s", token)
	return nil
}

func (r *SubscriptionRepo) IsSubscriptionExists(ctx context.Context, email, city string) (bool, error) {
	r.logger.Debugf("Checking if subscription exists for email: %s, city: %s", email, city)

	var count int64
	err := r.db.WithContext(ctx).Model(&Subscription{}).
		Where("email = ? AND city = ?", email, city).
		Count(&count).Error

	if err != nil {
		r.logger.Errorf("Failed to check subscription existence: %v", err)
		return false, err
	}

	exists := count > 0
	r.logger.Debugf("Subscription exists check for email %s, city %s: %t", email, city, exists)
	return exists, nil
}

func (r *SubscriptionRepo) ListByFrequency(ctx context.Context, frequency string, lastID, pageSize int) ([]domain.Subscription, error) {
	r.logger.Debugf("Listing subscriptions by frequency: %s, lastID: %d, pageSize: %d", frequency, lastID, pageSize)

	var subs []Subscription

	query := r.db.WithContext(ctx).Where("frequency = ? AND confirmed = ?", frequency, true)

	if lastID > 0 {
		query = query.Where("id > ?", lastID)
	}

	if err := query.Limit(pageSize).Find(&subs).Error; err != nil {
		r.logger.Errorf("Failed to list subscriptions by frequency %s: %v", frequency, err)
		return nil, err
	}

	result := make([]domain.Subscription, len(subs))
	for i, sub := range subs {
		result[i] = domain.Subscription{
			ID:          int64(sub.ID),
			Email:       sub.Email,
			City:        sub.City,
			Frequency:   string(sub.Frequency),
			Token:       sub.Token,
			IsConfirmed: sub.Confirmed,
		}
	}

	r.logger.Infof("Found %d subscriptions for frequency: %s", len(result), frequency)
	return result, nil
}
