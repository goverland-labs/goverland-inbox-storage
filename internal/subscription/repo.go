package subscription

import (
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(item UserSubscription) error {
	return r.db.Create(&item).Error
}

func (r *Repo) Delete(item UserSubscription) error {
	return r.db.Delete(&item).Error
}

func (r *Repo) GetByID(subscriberID, daoID string) (UserSubscription, error) {
	var res UserSubscription
	err := r.db.
		Where(&UserSubscription{
			UserID: subscriberID,
			DaoID:  daoID,
		}).
		First(&res).
		Error

	return res, err
}

// todo: think about getting this elements by chunks
func (r *Repo) GetSubscribers(daoID string) ([]UserSubscription, error) {
	var res []UserSubscription
	err := r.db.
		Where(&UserSubscription{
			DaoID: daoID,
		}).
		Find(&res).
		Error

	return res, err
}
