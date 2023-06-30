package subscription

import (
	"fmt"

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

func (r *Repo) GetBySubscriberAndDaoID(subscriberID, daoID string) (UserSubscription, error) {
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

func (r *Repo) GetByID(id string) (*UserSubscription, error) {
	us := UserSubscription{ID: id}
	request := r.db.Take(&us)
	if err := request.Error; err != nil {
		return nil, fmt.Errorf("get user subscription by id #%s: %w", id, err)
	}

	return &us, nil
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

func (r *Repo) GetByFilters(filters []Filter) (UserSubscriptionList, error) {
	db := r.db.Model(&UserSubscription{})
	for _, f := range filters {
		if _, ok := f.(PageFilter); ok {
			continue
		}
		db = f.Apply(db)
	}

	var cnt int64
	err := db.Count(&cnt).Error
	if err != nil {
		return UserSubscriptionList{}, err
	}

	for _, f := range filters {
		if _, ok := f.(PageFilter); ok {
			db = f.Apply(db)
		}
	}

	var list []UserSubscription
	err = db.Find(&list).Error
	if err != nil {
		return UserSubscriptionList{}, err
	}

	return UserSubscriptionList{
		Subscriptions: list,
		TotalCount:    cnt,
	}, nil
}
