package subscription

import (
	"time"

	"gorm.io/gorm"
)

type UserSubscription struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UserID    string
	DaoID     string
}

type GlobalSubscription struct {
	ID           string `gorm:"primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	SubscriberID string
	DaoID        string
}

type UserSubscriptionList struct {
	Subscriptions []UserSubscription
	TotalCount    int64
}
