package proposal

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type FeaturedRepo struct {
	db *gorm.DB
}

func NewFeaturedRepo(db *gorm.DB) *FeaturedRepo {
	return &FeaturedRepo{
		db: db,
	}
}

func (r *FeaturedRepo) GetFeaturedProposals(date time.Time) ([]Featured, error) {
	var featured []Featured

	err := r.db.
		Where("start_at <= @date and end_at > @date", sql.Named("date", date)).
		Find(&featured).
		Error
	if err != nil {
		return nil, err
	}

	return featured, nil
}
