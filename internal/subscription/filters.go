package subscription

import (
	"gorm.io/gorm"
)

type Filter interface {
	Apply(*gorm.DB) *gorm.DB
}

type PageFilter struct {
	Offset int
	Limit  int
}

func (f PageFilter) Apply(db *gorm.DB) *gorm.DB {
	return db.Offset(f.Offset).Limit(f.Limit)
}

type UserIDFilter struct {
	ID string
}

func (f UserIDFilter) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("user_id = ?", f.ID)
}
