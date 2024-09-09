package delegate

import (
	"time"

	"gorm.io/gorm"
)

type AllowedDao struct {
	DaoName   string
	CreatedAt time.Time
}

func (a *AllowedDao) TableName() string {
	return "delegate_allowed_daos"
}

type AllowedDaoRepo struct {
	db *gorm.DB
}

func NewAllowedDaoRepo(db *gorm.DB) *AllowedDaoRepo {
	return &AllowedDaoRepo{db: db}
}

func (r *AllowedDaoRepo) List() ([]AllowedDao, error) {
	var list []AllowedDao
	request := r.db.Find(&list)
	if err := request.Error; err != nil {
		return nil, err
	}

	return list, nil
}
