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

	UnknownRole Role = ""
	GuestRole   Role = "GUEST"
	RegularRole Role = "REGULAR"
)

type Role string

type User struct {
	ID uuid.UUID `gorm:"primary_key"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Role Role

	Address    *string
	ENS        *string
	DeviceUUID string // only for guest support, remove in future
}

type Session struct {
	ID     uuid.UUID `gorm:"primary_key"`
	UserID uuid.UUID `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	DeviceUUID string
	DeviceName string
}

func (s *Session) TableName() string {
	return "user_sessions"
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
