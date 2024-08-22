package appversions

import (
	"gorm.io/gorm"
)

type Filter interface {
	Apply(*gorm.DB) *gorm.DB
}

type PlatformFilter struct {
	Platform Platform
}

func (f PlatformFilter) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("platform = ?", f.Platform)
}
