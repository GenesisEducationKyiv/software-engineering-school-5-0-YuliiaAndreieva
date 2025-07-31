package database

import "time"

type Frequency string

type Subscription struct {
	ID        int    `gorm:"primaryKey"`
	Email     string `gorm:"unique"`
	City      string
	Token     string `gorm:"unique"`
	Frequency Frequency
	Confirmed bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

const (
	Daily  Frequency = "daily"
	Hourly Frequency = "hourly"
)
