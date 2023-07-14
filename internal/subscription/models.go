package subscription

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSubscription struct {
	ID        uuid.UUID `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UserID    uuid.UUID
	DaoID     uuid.UUID
}

type GlobalSubscription struct {
	ID           uuid.UUID `gorm:"primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	SubscriberID uuid.UUID
	DaoID        uuid.UUID
}

type UserSubscriptionList struct {
	Subscriptions []UserSubscription
	TotalCount    int64
}
