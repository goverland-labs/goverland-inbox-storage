package user

import (
	"time"
)

type User struct {
	ID         string `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
	DeviceUUID string
}
