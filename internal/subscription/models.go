package subscription

import "gorm.io/gorm"

type UserSubscription struct {
	gorm.Model
	UserID string
	DaoID  string
}

type GlobalSubscription struct {
	gorm.Model
	SubscriberID string
	DaoID        string
}
