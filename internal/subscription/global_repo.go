package subscription

import (
	"gorm.io/gorm"
)

type GlobalRepo struct {
	db *gorm.DB
}

func NewGlobalRepo(db *gorm.DB) *GlobalRepo {
	return &GlobalRepo{db: db}
}

func (r *GlobalRepo) Create(item GlobalSubscription) error {
	return r.db.Create(&item).Error
}

func (r *GlobalRepo) Delete(item GlobalSubscription) error {
	return r.db.Delete(&item).Error
}

func (r *GlobalRepo) GetByID(subscriberID, daoID string) (GlobalSubscription, error) {
	var res GlobalSubscription
	err := r.db.
		Where(&GlobalSubscription{
			SubscriberID: subscriberID,
			DaoID:        daoID,
		}).
		First(&res).
		Error

	return res, err
}
