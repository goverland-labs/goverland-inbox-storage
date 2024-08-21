package appversions

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Platform string

const (
	PlatformIos     Platform = "iOS"
	PlatformAndroid Platform = "android"
)

type Info struct {
	Version     string
	CreatedAt   time.Time
	Platform    Platform
	Description string
}

func (Info) TableName() string {
	return "app_versions"
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) GetByFilters(filters []Filter) ([]Info, error) {
	db := r.db.Model(&Info{})
	for _, f := range filters {
		db = f.Apply(db)
	}

	var list []Info
	err := db.Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	return list, nil
}
