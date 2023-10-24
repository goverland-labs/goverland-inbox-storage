package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecentlyType string

const (
	RecentlyTypeUnspecified = "unspecified"
	RecentlyTypeDao         = "dao"
)

type User struct {
	ID         string `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	DeviceUUID string
}

type RecentlyViewed struct {
	gorm.Model

	UserID uuid.UUID
	Type   RecentlyType
	TypeID string
}

func (rv *RecentlyViewed) TableName() string {
	return "recently_viewed"
}

type RecentlyViewedList struct {
	Views      []RecentlyViewed
	TotalCount int64
}
