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
	return &SubscriptionRepo{
		db:     db,
		logger: logger,
	}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s domain.Subscription) error {
	r.logger.Debugf("Creating subscription in database for email: %s, city: %s", s.Email, s.City)

	var existingCount int64
	err := r.db.WithContext(ctx).Model(&Subscription{}).
		Where("email = ? AND city = ? AND frequency = ?", s.Email, s.City, s.Frequency).
		Count(&existingCount).Error

	if err != nil {
		r.logger.Errorf("Failed to check existing subscription: %v", err)
		return err
	}

	if existingCount > 0 {
		r.logger.Warnf("Subscription already exists for email: %s, city: %s, frequency: %s", s.Email, s.City, s.Frequency)
		return domain.ErrDuplicateSubscription
	}

	sub := Subscription{
		Email:     s.Email,
		City:      s.City,
		Token:     s.Token,
		Frequency: Frequency(s.Frequency),
		Confirmed: s.IsConfirmed,
	}

	err = r.db.WithContext(ctx).Create(&sub).Error
	if err != nil {
		r.logger.Errorf("Failed to create subscription in database: %v", err)
		return err
	}

	r.logger.Infof("Successfully created subscription in database for email: %s, city: %s", s.Email, s.City)
	return nil
}

func (r *SubscriptionRepo) GetByToken(ctx context.Context, token string) (*domain.Subscription, error) {
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

func (r *SubscriptionRepo) Update(ctx context.Context, s domain.Subscription) error {
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

func (r *SubscriptionRepo) Delete(ctx context.Context, token string) error {
	r.logger.Debugf("Deleting subscription for token: %s", token)

	err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&Subscription{}).Error
	if err != nil {
		r.logger.Errorf("Failed to delete subscription for token %s: %v", token, err)
		return err
	}

	r.logger.Infof("Successfully deleted subscription for token: %s", token)
	return nil
}

func (r *SubscriptionRepo) ListByFrequency(ctx context.Context, frequency domain.Frequency, lastID, pageSize int) (*domain.SubscriptionList, error) {
	r.logger.Debugf("Listing subscriptions by frequency: %s, lastID: %d, pageSize: %d", frequency, lastID, pageSize)

	var subs []Subscription

	query := r.db.WithContext(ctx).Where("frequency = ? AND confirmed = ?", string(frequency), true)

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

	lastIndex := 0
	if len(result) > 0 {
		lastIndex = int(result[len(result)-1].ID)
	}

	r.logger.Infof("Found %d subscriptions for frequency: %s", len(result), frequency)
	return &domain.SubscriptionList{
		Subscriptions: result,
		LastIndex:     lastIndex,
	}, nil
}
