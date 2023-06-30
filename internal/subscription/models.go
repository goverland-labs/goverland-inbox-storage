package subscription

import (
	"time"
)

type UserSubscription struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	UserID    string
	DaoID     string
}

type GlobalSubscription struct {
	ID           string `gorm:"primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    time.Time
	SubscriberID string
	DaoID        string
}

type UserSubscriptionList struct {
	Subscriptions []UserSubscription
	TotalCount    int64
}
