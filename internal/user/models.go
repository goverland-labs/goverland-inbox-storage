package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         string `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	DeviceUUID string
}
