package database

import (
	"context"
	"subscription-service/internal/core/domain"
	"subscription-service/internal/core/ports/out"

	"gorm.io/gorm"
)

type SubscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) out.SubscriptionRepository {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) CreateSubscription(ctx context.Context, s domain.Subscription) error {
	sub := Subscription{
		Email:     s.Email,
		City:      s.City,
		Token:     s.Token,
		Frequency: Frequency(s.Frequency),
		Confirmed: s.IsConfirmed,
	}
	return r.db.WithContext(ctx).Create(&sub).Error
}

func (r *SubscriptionRepo) GetSubscriptionByToken(ctx context.Context, token string) (*domain.Subscription, error) {
	var sub Subscription
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&sub).Error; err != nil {
		return nil, err
	}
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
	return r.db.WithContext(ctx).Model(&Subscription{}).
		Where("token = ?", s.Token).
		Updates(map[string]interface{}{
			"confirmed": s.IsConfirmed,
		}).Error
}

func (r *SubscriptionRepo) DeleteSubscription(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&Subscription{}).Error
}

func (r *SubscriptionRepo) IsSubscriptionExists(ctx context.Context, email, city string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Subscription{}).
		Where("email = ? AND city = ?", email, city).
		Count(&count).Error
	return count > 0, err
} 