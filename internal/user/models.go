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

func (u User) IsGuest() bool {
	return u.Role == GuestRole
}

func (u User) IsRegular() bool {
	return u.Role == RegularRole
}

func (u User) HasAddress() bool {
	return u.Address != nil && *u.Address != ""
}

type Session struct {
	ID     uuid.UUID `gorm:"primary_key"`
	UserID uuid.UUID `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	DeviceUUID string
	DeviceName string
	AppVersion string

	LastActivityAt time.Time
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

type Activity struct {
	gorm.Model

	UserID     uuid.UUID
	FinishedAt time.Time
}

func (a *Activity) TableName() string {
	return "user_activity"
}

type AuthNonce struct {
	Address   string    `gorm:"primary_key"`
	Nonce     string    `gorm:"primary_key"`
	ExpiredAt time.Time `gorm:"primary_key"`
}

func (a *AuthNonce) TableName() string {
	return "auth_nonces"
}

type CanVote struct {
	UserID     uuid.UUID `gorm:"primary_key"`
	ProposalID string    `gorm:"primary_key"`

	CreatedAt time.Time `gorm:"index"`
}

func (u *CanVote) TableName() string {
	return "user_can_vote"
}
